# Sidecar Injection

An attacker with permissions to modify pod specs can inject a malicious sidecar container into a running workload. Because all containers in a pod share the same network namespace, the sidecar has full visibility into unencrypted traffic flowing through the application — without touching the application image at all.

## Description

A sidecar is a container that runs alongside the main container in a pod, sharing its network and storage namespaces. Legitimate sidecars add logging, proxying, or monitoring. An attacker who has gained `patch` or `update` permissions on Deployments can inject a hostile sidecar that:

- Sniffs unencrypted HTTP traffic with `tcpdump` and exfiltrates it to an external endpoint.
- Reads files from shared volumes (secrets, configs, tokens).
- Provides a persistent reverse shell inside an otherwise legitimate workload.

This technique is stealthy because the injected container runs under the existing pod's identity and the application workload continues functioning normally.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- The attacker has obtained `patch` access to Deployments in the target namespace (e.g., via a stolen kubeconfig or a misconfigured RBAC role).

## Quick Start

### Step 1 — Deploy the target application

Deploy an nginx-fronted FastAPI application that will be the injection target.

```bash
kubectl apply -f nginx.yaml
```

Wait for the pod to become ready:

```bash
kubectl rollout status deployment/nginx
```

Expected output:

```
deployment.apps/nginx successfully rolled out
```

Verify the service is reachable from within the cluster:

```bash
kubectl run curl-test --image=curlimages/curl:latest --restart=Never --rm -it -- \
  curl -s http://nginx:8080/
```

### Step 2 — Inspect the sidecar patch manifest

The file `sidecard-injection.yaml` is a strategic merge patch that adds a privileged `snooper` container to the existing pod template. The sidecar:

1. Installs `tcpdump` and `curl` at runtime (alpine-based, no custom image needed).
2. Captures all HTTP traffic on any interface on port 80.
3. Pipes captured packets through `awk` and exfiltrates each HTTP request to an external webhook.

Review the patch before applying:

```bash
cat sidecard-injection.yaml
```

### Step 3 — Inject the sidecar

Apply the strategic merge patch to the `nginx` Deployment:

```bash
kubectl patch deployment nginx --patch-file sidecard-injection.yaml
```

Expected output:

```
deployment.apps/nginx patched
```

Wait for the rollout to complete with the new sidecar:

```bash
kubectl rollout status deployment/nginx
```

Confirm both containers are running in the pod:

```bash
kubectl get pods -l app=nginx -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{range .spec.containers[*]}{.name}{" "}{end}{"\n"}{end}'
```

Expected output:

```
nginx-6d8f9b7c4-xk9pl    snooper nginx
```

### Step 4 — Generate traffic and observe data exfiltration

Send several HTTP requests to the application service to produce traffic for the sidecar to capture:

```bash
for i in $(seq 1 5); do
  kubectl run curl-traffic-$i --image=curlimages/curl:latest --restart=Never --rm -it -- \
    curl -s -H "Authorization: Bearer supersecret-token-$i" http://nginx:8080/
done
```

Watch the snooper sidecar's logs to observe captured packets in real time:

```bash
kubectl logs -l app=nginx -c snooper --follow
```

The snooper is forwarding every captured HTTP request — including headers containing the `Authorization` bearer tokens — to the external webhook endpoint defined in `sidecard-injection.yaml`. In a real attack this webhook would be the attacker's collection server.

### Step 5 — Verify the sidecar persists through pod restarts

Delete the pod manually to simulate a crash or node eviction:

```bash
kubectl delete pod -l app=nginx
```

The Deployment controller immediately schedules a replacement pod. Confirm the sidecar is still present in the new pod:

```bash
kubectl rollout status deployment/nginx
kubectl get pods -l app=nginx -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{range .spec.containers[*]}{.name}{" "}{end}{"\n"}{end}'
```

The sidecar persists because the patch was applied to the Deployment spec, not just to a single pod instance.

## Cleanup

```bash
kubectl delete -f nginx.yaml
```

## Resources

- [Sidecar Pattern](https://kubernetes.io/docs/concepts/workloads/pods/sidecar-containers/)
- [kubectl patch](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#patch)
- [Strategic Merge Patch](https://kubernetes.io/docs/tasks/manage-kubernetes-objects/update-api-object-kubectl-patch/#use-a-strategic-merge-patch-to-update-a-deployment)
- [MITRE ATT&CK for Containers — Sidecar Injection](https://microsoft.github.io/Threat-Matrix-for-Kubernetes/techniques/Sidecar%20injection/)
