# Network Mapping

An attacker who gains a foothold inside a pod can treat that pod as a pivot point to enumerate the rest of the cluster. Without egress restrictions, standard Linux networking tools are sufficient to map every service, pod, and node reachable from within the pod network.

## Description

Attackers may use network scanning tools such as `nmap` or `zmap` to map the cluster's network. After gaining access to a container, an attacker can query the Kubernetes DNS service (CoreDNS) to resolve service names, enumerate listening ports across the entire pod CIDR and service CIDR, and identify vulnerable or misconfigured applications running elsewhere in the cluster. This reconnaissance phase is typically a precursor to lateral movement.

## Prerequisites

- A running Kind cluster (`workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.

## Quick Start

### Step 1 — Deploy the scenario

Deploy a set of services across multiple namespaces to simulate a realistic multi-tenant environment, and deploy an attacker pod that contains network scanning tools.

```bash
kubectl apply -f scenario.yaml
```

Wait for all pods to become ready:

```bash
kubectl wait --for=condition=Ready pod -l app=nginx -n web-apps --timeout=60s
kubectl wait --for=condition=Ready pod -l role=attacker -n attacker --timeout=60s
```

Verify the resources:

```bash
kubectl get all -n web-apps
kubectl get all -n internal-api
kubectl get all -n attacker
```

Example output:

```
NAME                        READY   STATUS    RESTARTS   AGE
pod/frontend-...            1/1     Running   0          30s
pod/backend-...             1/1     Running   0          30s

NAME               TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
service/frontend   ClusterIP   10.96.100.10    <none>        80/TCP     30s
service/backend    ClusterIP   10.96.200.20    <none>        8080/TCP   30s
```

### Step 2 — Exec into the attacker pod

```bash
kubectl exec -it -n attacker deploy/attacker -- sh
```

All subsequent commands in this section run **inside the attacker pod**.

### Step 3 — DNS-based service discovery

CoreDNS resolves all in-cluster services. Query it directly to enumerate known service names:

```bash
# Resolve services by their fully-qualified domain names
nslookup frontend.web-apps.svc.cluster.local
nslookup backend.web-apps.svc.cluster.local
nslookup private-api.internal-api.svc.cluster.local
nslookup kubernetes.default.svc.cluster.local
```

Example output:

```
Server:         10.96.0.10
Address:        10.96.0.10#53

Name:   frontend.web-apps.svc.cluster.local
Address: 10.107.13.48
```

Discover the DNS search domain and service CIDR hint from `/etc/resolv.conf`:

```bash
cat /etc/resolv.conf
```

Example output:

```
search attacker.svc.cluster.local svc.cluster.local cluster.local
nameserver 10.96.0.10
options ndots:5
```

### Step 4 — Enumerate the Kubernetes API server

The API server address is injected as an environment variable into every pod:

```bash
env | grep -i kubernetes
curl -sk https://${KUBERNETES_SERVICE_HOST}:${KUBERNETES_SERVICE_PORT}/version
```

Example output:

```
{
  "major": "1",
  "minor": "30",
  "gitVersion": "v1.30.2",
  ...
}
```

### Step 5 — Scan the service CIDR range

Identify the service CIDR from the cluster info and scan it with `nmap`. In Kind clusters the service CIDR is `10.96.0.0/12` but scanning the full /12 is slow; use the /24 containing the nameserver address as a starting point:

```bash
# The service CIDR is typically printed in the resolv.conf nameserver or discoverable via:
# Scan port 80, 443, 8080, 8443 across the service subnet (use /24 for speed)
nmap -sT -p 80,443,8080,8443,9090,9093,9200,6379,5432,3306 \
     --open -T4 10.96.0.0/24 2>/dev/null
```

Example output:

```
Nmap scan report for kubernetes.default.svc.cluster.local (10.96.0.1)
Host is up (0.000028s latency).
PORT    STATE SERVICE
443/tcp open  https

Nmap scan report for frontend.web-apps.svc.cluster.local (10.107.13.48)
Host is up (0.000023s latency).
PORT   STATE SERVICE
80/tcp open  http

Nmap scan report for backend.web-apps.svc.cluster.local (10.99.161.223)
Host is up (0.00013s latency).
PORT     STATE SERVICE
8080/tcp open  http-proxy
...
```

### Step 6 — Scan the pod CIDR range

Pod IPs are allocated from a separate CIDR (typically `10.244.0.0/16` in Kind clusters). Scan for live hosts and common application ports:

```bash
# Identify the pod's own IP to determine the CIDR
ip addr show eth0

# Scan the pod network for live hosts and open ports
nmap -sT -p 80,443,8080,8443,8000,3000,9090 \
     --open -T4 10.244.0.0/16 2>/dev/null
```

### Step 7 — Identify running services and reach across namespaces

With the discovered IP addresses or DNS names, probe services directly:

```bash
# Reach the frontend service in the web-apps namespace
curl -s http://frontend.web-apps.svc.cluster.local/

# Reach the private API in the internal-api namespace
curl -s http://private-api.internal-api.svc.cluster.local:8080/

# Attempt to reach the Kubernetes API (will fail without a valid token)
curl -sk https://kubernetes.default.svc.cluster.local/api/v1/namespaces
```

### Step 8 — Scan kubelet ports on cluster nodes

Kubelet exposes a management API on port 10250 and optionally a read-only API on port 10255 (disabled by default since Kubernetes 1.16):

```bash
# Discover node IPs — they are typically in a different subnet (e.g. 172.23.0.0/24 in Kind)
# The exact subnet varies by Kind configuration; derive it from the node's gateway or
# by inspecting the attacker pod's default route.
# Scan for kubelet and etcd ports
nmap -sT -p 10250,10255,2379,2380 --open -T4 172.23.0.0/24 2>/dev/null
```

Example output (Kind cluster with one control-plane and three workers):

```
Nmap scan report for kind-worker2.kind (172.23.0.2)
Host is up (0.000049s latency).
PORT      STATE SERVICE
10250/tcp open  unknown

Nmap scan report for kind-worker.kind (172.23.0.3)
Host is up (0.000026s latency).
PORT      STATE SERVICE
10250/tcp open  unknown

Nmap scan report for 172-23-0-4.kubernetes.default.svc.cluster.local (172.23.0.4)
Host is up (0.000047s latency).
PORT      STATE SERVICE
2379/tcp  open  etcd-client
2380/tcp  open  etcd-server
10250/tcp open  unknown

Nmap scan report for kind-worker3.kind (172.23.0.5)
Host is up (0.000052s latency).
PORT      STATE SERVICE
10250/tcp open  unknown
```

> **Note:** Port 10255 (kubelet read-only API) is disabled by default since Kubernetes 1.16. The node subnet (`172.23.0.0/24` above) depends on your Kind network configuration and will differ across environments.

Exit the attacker pod when done:

```bash
exit
```

## Cleanup

```bash
kubectl delete -f scenario.yaml
```

## Resources

- [nmap](https://nmap.org/)
- [zmap](https://zmap.io/)
- [MITRE ATT&CK - Network Service Discovery](https://attack.mitre.org/techniques/T1046/)
- [Kubernetes DNS for Services and Pods](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/)
- [Kubernetes Network Model](https://kubernetes.io/docs/concepts/cluster-administration/networking/)
