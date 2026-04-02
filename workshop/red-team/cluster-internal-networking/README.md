# Cluster Internal Networking

By default, Kubernetes places no restrictions on which pods can communicate with which other pods. An attacker who compromises a single container can immediately reach every other service in the cluster — across namespaces — using standard HTTP clients. This is lateral movement with zero additional exploitation required.

## Description

Kubernetes networking behavior allows traffic between pods in the cluster as a default behavior. Attackers who gain access to a single container may use it for network reachability to another container in the cluster. Without explicit NetworkPolicy objects (enforced by a CNI plugin that supports them, such as Calico), every pod is on a flat network with unrestricted east-west connectivity.

This lab demonstrates:

1. **Open by default** — cross-namespace service access works out of the box.
2. **Attacker perspective** — a breached pod in `tenant-1` can reach services in `tenant-2` and `tenant-3`.
3. **Remediation** — Calico NetworkPolicy objects progressively restrict access.

## Prerequisites

- A running Kind cluster (`workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- `k9s` (optional, for interactive pod exec).

> **Important — Calico and Kind:** Steps 3-8 require Calico to enforce NetworkPolicy. Calico's data-plane enforcement **does not work** when the cluster uses the default `kindnet` CNI, because `kindnet` controls pod routing and Calico's WorkloadEndpoints are never registered. To use Calico, the Kind cluster must be created **without** the default CNI:
>
> ```bash
> # kind-config.yaml
> kind: Cluster
> apiVersion: kind.x-k8s.io/v1alpha4
> networking:
>   disableDefaultCNI: true
>   podSubnet: "10.244.0.0/16"
> nodes:
>   - role: control-plane
>   - role: worker
>   - role: worker
>   - role: worker
> ```
> ```bash
> kind create cluster --config kind-config.yaml --name workshop-cluster
> ```
>
> Steps 1-2 (unrestricted cross-namespace access) work with any CNI and can be tested on a standard Kind cluster.

## Quick Start

### Step 1 — Deploy tenant workloads

Each tenant YAML deploys a Namespace, a ConfigMap with an nginx configuration, a Deployment, and a Service. The tenants simulate isolated application teams sharing the same cluster.

```bash
kubectl apply -f tenant-1.yaml
kubectl apply -f tenant-2.yaml
```

Wait for pods to become ready:

```bash
kubectl wait --for=condition=Ready pod -l app=nginx -n tenant-1 --timeout=60s
kubectl wait --for=condition=Ready pod -l app=nginx -n tenant-2 --timeout=60s
```

Inspect what was deployed:

```bash
kubectl get all --namespace tenant-1
kubectl get all --namespace tenant-2
```

Example output:

```
# tenant-1
NAME                         READY   STATUS    RESTARTS   AGE
pod/nginx-6d4cf56db6-xk2p9   1/1     Running   0          30s

NAME            TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
service/nginx   ClusterIP   10.96.10.100   <none>        8080/TCP   30s
```

### Step 2 — Demonstrate unrestricted cross-namespace access (attacker perspective)

Exec into the nginx pod in `tenant-1`. This simulates an attacker who has already compromised the tenant-1 workload.

```bash
# Get the pod name
TENANT1_POD=$(kubectl get pod -n tenant-1 -l app=nginx -o jsonpath='{.items[0].metadata.name}')

# Exec into the compromised container
kubectl exec -it -n tenant-1 ${TENANT1_POD} -- sh
```

Inside the pod, install `curl` and reach the `tenant-2` service:

```bash
# Install curl (Alpine-based image)
apk add --no-cache curl

# Reach tenant-2's nginx service — different namespace, no restrictions
curl -s http://nginx.tenant-2.svc.cluster.local:8080
```

Expected output — the tenant-2 web page is returned:

```html
<!DOCTYPE html>
<html>
...
    <h1>Tenant Two</h1>
...
</html>
```

Kubernetes service DNS follows the pattern: `<service>.<namespace>.svc.cluster.local:<port>`

```bash
# Also reachable by ClusterIP directly
curl -s http://nginx.tenant-2.svc.cluster.local:8080
exit
```

Now exec into the `tenant-2` pod and confirm it can reach `tenant-1`:

```bash
TENANT2_POD=$(kubectl get pod -n tenant-2 -l app=nginx -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it -n tenant-2 ${TENANT2_POD} -- sh
```

```bash
apk add --no-cache curl

# Reach tenant-1 from tenant-2
curl -s http://nginx.tenant-1.svc.cluster.local:8080
exit
```

Both tenants can reach each other with no authentication or authorization required.

### Step 3 — Install Calico to enforce NetworkPolicy

Standard Kubernetes NetworkPolicy objects require a CNI plugin that enforces them. Calico is one such plugin. Install the Tigera operator first:

```bash
kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.1/manifests/tigera-operator.yaml
```

Install Calico by creating the custom resource. Review the IP pool CIDR — it must match your cluster's pod CIDR (`10.244.0.0/16` is the Kind default):

```bash
kubectl create -f custom-resources.yaml
```

Wait for the Calico system pods to become ready:

```bash
watch kubectl get pods -n calico-system
```

Wait until all pods show `Running`:

```
NAME                                       READY   STATUS    RESTARTS   AGE
calico-kube-controllers-...               1/1     Running   0          60s
calico-node-xxxxx                          1/1     Running   0          60s
calico-typha-...                           1/1     Running   0          60s
```

### Step 4 — Apply a default-deny policy to tenant-2

The `np-default-deny.yaml` policy blocks all ingress traffic to the `tenant-2` namespace. It uses Calico's `NetworkPolicy` CRD with `order: 20` (higher order = lower precedence; allow rules at `order: 10` will override this when needed).

```bash
kubectl apply -f np-default-deny.yaml
```

> **Note:** In some cluster setups nginx pods may need to be restarted after Calico installs its eBPF or iptables rules. If pods appear stuck, run: `kubectl rollout restart deployment/nginx -n tenant-2`

### Step 5 — Verify tenant-1 can no longer reach tenant-2

Exec into tenant-1 again and try reaching tenant-2:

```bash
TENANT1_POD=$(kubectl get pod -n tenant-1 -l app=nginx -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it -n tenant-1 ${TENANT1_POD} -- sh
```

```bash
# Install curl if not already installed
apk add --no-cache curl

# This should now time out — lateral movement is blocked
curl -m 5 http://nginx.tenant-2.svc.cluster.local:8080
exit
```

Expected output:

```
curl: (28) Connection timed out after 5001 milliseconds
```

The attacker's lateral movement path is closed.

### Step 6 — Verify tenant-2 internal traffic is also blocked

Exec into tenant-2 and test different connectivity scenarios:

```bash
TENANT2_POD=$(kubectl get pod -n tenant-2 -l app=nginx -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it -n tenant-2 ${TENANT2_POD} -- sh
```

```bash
apk add --no-cache curl

# Cross-namespace to tenant-1 — still works (we only blocked ingress to tenant-2)
curl -m 5 http://nginx.tenant-1.svc.cluster.local:8080

# Intra-namespace via service name — times out (blocked by default-deny on tenant-2)
curl -m 5 http://nginx.tenant-2.svc.cluster.local:8080

# Localhost — always works (traffic doesn't traverse the network policy)
curl -m 5 http://localhost:8080
exit
```

### Step 7 — Allow intra-namespace traffic for tenant-2

The `np-allow-namespace-connectivity.yaml` policy adds an allow rule at `order: 10` (higher precedence than the deny at `order: 20`) to permit ingress from within `tenant-2` itself:

```bash
kubectl apply -f np-allow-namespace-connectivity.yaml
```

Exec into tenant-2 and verify internal connectivity is restored while tenant-1 is still blocked:

```bash
kubectl exec -it -n tenant-2 ${TENANT2_POD} -- sh
```

```bash
# Intra-namespace via service name — now works
curl -m 5 http://nginx.tenant-2.svc.cluster.local:8080

# tenant-1 is still blocked from reaching tenant-2 (ingress from tenant-1 not allowed)
exit
```

### Step 8 — Deploy tenant-3 and grant selective cross-namespace access

Deploy a third tenant and update the tenant-2 policy to allow ingress from `tenant-3`:

```bash
kubectl apply -f tenant-3.yaml
kubectl wait --for=condition=Ready pod -l app=nginx -n tenant-3 --timeout=60s

# Update the NetworkPolicy to additionally allow ingress from tenant-3
kubectl apply -f np-allow-namespace-connectivity-update.yaml
```

Exec into tenant-3 and verify it can reach tenant-2, while tenant-1 still cannot:

```bash
TENANT3_POD=$(kubectl get pod -n tenant-3 -l app=nginx -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it -n tenant-3 ${TENANT3_POD} -- sh
```

```bash
apk add --no-cache curl

# tenant-3 can reach tenant-2 (explicitly allowed)
curl -m 5 http://nginx.tenant-2.svc.cluster.local:8080

# tenant-2 can be reached from tenant-3
exit
```

From tenant-1, confirm it is still blocked from tenant-2:

```bash
TENANT1_POD=$(kubectl get pod -n tenant-1 -l app=nginx -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it -n tenant-1 ${TENANT1_POD} -- sh -c \
    "curl -m 5 http://nginx.tenant-2.svc.cluster.local:8080 || echo BLOCKED"
```

## Cleanup

```bash
kubectl delete -f tenant-1.yaml
kubectl delete -f tenant-2.yaml
kubectl delete -f tenant-3.yaml
kubectl delete -f np-default-deny.yaml
kubectl delete -f np-allow-namespace-connectivity.yaml
kubectl delete -f np-allow-namespace-connectivity-update.yaml
kubectl delete -f custom-resources.yaml
kubectl delete -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.1/manifests/tigera-operator.yaml
```

## Resources

- [Kubernetes Networking](https://kubernetes.io/docs/concepts/cluster-administration/networking/)
- [Kubernetes NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [Calico Network Policy](https://docs.tigera.io/calico/latest/network-policy/get-started/calico-policy/calico-network-policy)
- [Calico Quickstart for Kubernetes](https://docs.tigera.io/calico/latest/getting-started/kubernetes/quickstart)
- [MITRE ATT&CK - Lateral Movement](https://attack.mitre.org/tactics/TA0008/)
- [MITRE ATT&CK - Internal Spearphishing](https://attack.mitre.org/techniques/T1534/)
