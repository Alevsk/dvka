# Connect from Proxy Server

An attacker who has gained a foothold inside a Kubernetes cluster can use a compromised pod as a network pivot point — proxying traffic through it to reach internal cluster services, the Kubernetes API server, or other network segments that are not directly accessible from the attacker's external machine.

## Description

Attackers may use proxy servers to hide their origin IP. Specifically, attackers often use anonymous networks such as TOR for their activity. This can be used for communicating with the applications themselves or with the API server.

Inside a Kubernetes cluster, a compromised pod provides a natural pivot: it has a cluster-internal IP, access to the cluster DNS, and can often reach services that are not exposed externally. Attackers can use several techniques to establish a proxy through a compromised pod:

- **`kubectl port-forward`**: Tunnel a local port to a port inside a pod over the existing kubectl connection.
- **`kubectl proxy`**: Start a local HTTP proxy to the Kubernetes API server, forwarding requests as the current kubeconfig user.
- **In-pod SOCKS proxy**: Deploy a SOCKS5 proxy inside a compromised pod and use it as a pivot for scanning or accessing internal services.

## Prerequisites

- A running Kubernetes cluster (e.g., `workshop-cluster` via Kind).
- `kubectl` installed and configured to connect to your cluster.

## Quick Start

### 1. Deploy the pivot pod

The pivot pod runs a simple web server on port 8080 and also includes `curl` for making internal requests. In a real attack scenario this would be any compromised workload.

```bash
kubectl apply -f pivot-pod.yaml
```

Wait for the pod to be running:

```bash
kubectl get pod pivot-pod -n default -w
```

### 2. Technique A — kubectl port-forward as a tunnel

`kubectl port-forward` opens a TCP tunnel from the attacker's local machine to a port inside the pod. This allows the attacker to access internal services as if they were running locally.

Forward local port 9090 to port 8080 inside the pivot pod:

```bash
kubectl port-forward pod/pivot-pod 9090:8080 -n default &
```

Now reach the pod's internal service from localhost:

```bash
curl -s http://localhost:9090
```

Expected output:

```html
<html><body><h1>pivot-pod internal service</h1></body></html>
```

This connection appears to the API server as a kubectl request, not as a direct network connection to the pod — hiding the attacker's true network origin.

### 3. Technique B — kubectl proxy to the API server

`kubectl proxy` starts a local HTTP server that proxies all requests to the Kubernetes API server, using the current kubeconfig credentials. An attacker with a stolen kubeconfig can open a proxy and interact with the API server without ever making a direct TLS connection to it.

```bash
kubectl proxy --port=8001 &
```

Interact with the Kubernetes API through the proxy:

```bash
# List all namespaces via the proxy
curl -s http://localhost:8001/api/v1/namespaces | python3 -m json.tool | grep '"name"'

# Retrieve all secrets in the default namespace
curl -s http://localhost:8001/api/v1/namespaces/default/secrets | python3 -m json.tool

# Access the API discovery endpoint
curl -s http://localhost:8001/apis
```

Stop the proxy when done:

```bash
kill %1
```

### 4. Technique C — Use the pivot pod to reach internal cluster services

Exec into the pivot pod and use it to scan or access services that are not reachable from outside the cluster.

First, deploy an internal service that is not exposed externally:

```bash
kubectl apply -f internal-service.yaml
```

Use non-interactive exec to run commands inside the pivot pod:

```bash
# Access the internal service by its DNS name
kubectl exec pivot-pod -n default -- curl -s http://internal-service.default.svc.cluster.local

# Reach the Kubernetes API server directly from inside the cluster
kubectl exec pivot-pod -n default -- sh -c \
  'curl -sk https://kubernetes.default.svc.cluster.local/api \
   -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)"'
```

### 5. Technique D — Deploy a SOCKS5 proxy inside the pod

An attacker can deploy a SOCKS5 proxy process inside a compromised pod, then forward a local port to it, creating a full SOCKS5 tunnel into the cluster network.

```bash
# Exec into the pivot pod and start a simple SOCKS proxy with ssh -D
# In practice, attackers use tools like chisel, goproxy, or microsocks

# Example: check what tools are available in the pod
kubectl exec pivot-pod -n default -- sh -c 'which nc curl 2>/dev/null; echo "Available tools listed above"'

# Inside the pod, start a SOCKS5 proxy on port 1080 using a pre-installed tool
# Example with ncat or a similar tool in real engagements:
# kubectl exec pivot-pod -n default -- microsocks -p 1080 &
```

Forward the SOCKS proxy port to localhost:

```bash
kubectl port-forward pod/pivot-pod 1080:1080 -n default &
```

Configure your browser or tooling to use `socks5://localhost:1080` to route all traffic through the cluster network.

## Cleanup

```bash
kubectl delete -f pivot-pod.yaml
kubectl delete -f internal-service.yaml

# Kill any background port-forward or proxy processes
kill $(lsof -ti:9090) 2>/dev/null || true
kill $(lsof -ti:8001) 2>/dev/null || true
kill $(lsof -ti:1080) 2>/dev/null || true
```

## Resources

- [TOR Project](https://www.torproject.org/)
- [kubectl port-forward](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_port-forward/)
- [kubectl proxy](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_proxy/)
- [Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [MITRE ATT&CK — Proxy](https://attack.mitre.org/techniques/T1090/)
