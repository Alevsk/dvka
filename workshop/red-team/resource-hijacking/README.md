# Resource Hijacking

An attacker who gains the ability to schedule workloads on a Kubernetes cluster can deploy pods that consume cluster compute resources for their own benefit — most commonly cryptocurrency mining. Because Kubernetes does not restrict what a pod can run, and because many clusters lack per-namespace resource quotas or workload anomaly detection, mining pods can run undetected for extended periods.

## Description

Resource hijacking (also called cryptojacking in the context of mining) involves deploying workloads that consume CPU, memory, GPU, or network bandwidth for the attacker's benefit rather than the cluster owner's. In Kubernetes clusters this typically looks like:

- A Deployment disguised with a legitimate-sounding name (`logger`, `metrics-agent`, `cache-warmer`) running a CPU-intensive process.
- No resource `limits` set, allowing the pod to consume all available CPU on the node.
- No resource `requests` set (or very low ones), so the scheduler places the pod on a node that appears to have spare capacity.
- The pod runs in a non-default namespace to avoid casual inspection.

The impact extends beyond the hijacked compute: legitimate workloads on the same node are starved of CPU, latency increases, and cloud cost anomalies appear on the billing dashboard.

Detection signals include: node CPU utilization near 100% with no corresponding business load increase, `kubectl top` showing unexpected high-CPU pods, and process names like `xmrig`, `minergate`, or `stress` visible via `kubectl exec -- ps`.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- `metrics-server` deployed in the cluster (required for `kubectl top`).
- The attacker has `create` permissions on Deployments in at least one namespace.

## Quick Start

### Step 1 — Deploy a legitimate workload with resource quotas

First deploy the legitimate production workload that will be impacted by the attack:

```bash
kubectl apply -f resource-quota.yaml
```

Verify the quota and the legitimate application:

```bash
kubectl get resourcequota -n production
kubectl rollout status deployment/legitimate-app -n production
```

Expected output:

```
NAME               AGE   REQUEST                                          LIMIT
production-quota   10s   requests.cpu: 200m/2, requests.memory: 256Mi/2Gi ...
```

### Step 2 — Deploy the disguised miner

The miner pod is deployed in a separate namespace under the name `logger` to blend in with normal cluster operations. It runs `stress --cpu 2` with no CPU limit, allowing it to consume all available CPU on the node:

```bash
kubectl apply -f miner-pod.yaml
```

Verify the pod is running:

```bash
kubectl get pods -n cryptominer
```

Expected output:

```
NAME                      READY   STATUS    RESTARTS   AGE
logger-xxxxxxxxxx-xxxxx   1/1     Running   0          10s
```

Confirm what process is actually running inside the "logger" pod:

```bash
MINER_POD=$(kubectl get pod -n cryptominer -l app=logger -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n cryptominer "$MINER_POD" -- ps aux
```

Expected output:

```
USER         PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND
root           1  0.0  0.0 155932  7916 ?        Ssl  04:45   0:00 stress --cpu 2 --timeout 86400
root          14  112  0.0 156072  6604 ?        Rl   04:45   0:05 stress --cpu 2 --timeout 86400
root          16  112  0.0 156072  6608 ?        Rl   04:45   0:05 stress --cpu 2 --timeout 86400
```

### Step 3 — Observe the resource consumption

Check node-level CPU consumption (requires metrics-server):

```bash
kubectl top nodes
```

Expected output — the node CPU is now substantially elevated:

```
NAME                        CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
workshop-cluster-control-plane   1980m    99%    512Mi           25%
```

Check pod-level CPU consumption:

```bash
kubectl top pods -n cryptominer
kubectl top pods --all-namespaces --sort-by=cpu | head -10
```

Expected output:

```
NAMESPACE     NAME                    CPU(cores)   MEMORY(bytes)
cryptominer   logger-xxxxxxxxxx       1950m        4Mi
```

The `logger` pod is consuming nearly 2 full CPU cores while the `legitimate-app` pods in the `production` namespace are being starved.

Verify the impact on the legitimate workload by checking if new pods in the `production` namespace are pending:

```bash
kubectl scale deployment legitimate-app --replicas=5 -n production
kubectl get pods -n production
```

Some pods may be `Pending` because the node has no remaining CPU capacity to satisfy their resource requests.

### Step 4 — Inspect the miner Deployment for forensics

Examine the Deployment spec to understand how the attacker disguised the workload:

```bash
kubectl get deployment logger -n cryptominer -o yaml
```

Key red flags:
- Image: `progrium/stress` (not a typical application image)
- No `limits` defined for CPU
- Namespace: `cryptominer` (not a business namespace)
- Command: `stress --cpu 2`

Check the image pull history to see when the miner was deployed:

```bash
kubectl describe pod -n cryptominer -l app=logger | grep -A5 "Events:"
```

### Step 5 — Simulate scale-out (multi-node hijacking)

In a real attack the miner would scale to every node. Simulate this:

```bash
# Scale to match the number of nodes in the cluster
NODE_COUNT=$(kubectl get nodes --no-headers | wc -l)
kubectl scale deployment logger --replicas="$NODE_COUNT" -n cryptominer
kubectl get pods -n cryptominer -o wide
```

Each pod lands on a different node, hijacking CPU cluster-wide.

## Cleanup

```bash
kubectl delete -f miner-pod.yaml
kubectl delete -f resource-quota.yaml
```

## Resources

- [Cryptojacking](https://www.crowdstrike.com/cybersecurity-101/cryptojacking/)
- [Kubernetes Resource Quotas](https://kubernetes.io/docs/concepts/policy/resource-quotas/)
- [kubectl top](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#top)
- [metrics-server](https://github.com/kubernetes-sigs/metrics-server)
- [MITRE ATT&CK for Containers — Resource Hijacking](https://attack.mitre.org/techniques/T1496/)
- [MITRE ATT&CK for Kubernetes — Resource Hijacking](https://microsoft.github.io/Threat-Matrix-for-Kubernetes/techniques/Resource%20hijacking/)
