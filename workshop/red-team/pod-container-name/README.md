# Pod / Container Name Similarity

An attacker who can create pods in a Kubernetes cluster may name their malicious pod and its container to closely mimic a legitimate system component (such as `coredns` or `kube-proxy`) to blend into the cluster's existing workload and evade detection during a manual review.

## Description

Attackers may give their pods and containers names that are similar to the names of other objects in the cluster. This can be used to hide their malicious activity from the cluster administrator. By matching not only the name but also the namespace, labels, and container name of a real system component, an attacker can make their pod appear in `kubectl get pods` listings alongside the legitimate workloads, making it easy to overlook during a quick inspection.

## Prerequisites

- A running Kubernetes cluster (e.g., `workshop-cluster` via Kind).
- `kubectl` installed and configured to connect to your cluster.

## Quick Start

### 1. Observe the legitimate CoreDNS pods

List the real CoreDNS pods to understand the naming convention used by the cluster:

```bash
kubectl get pods -n kube-system -l k8s-app=kube-dns
```

Example output:

```
NAME                       READY   STATUS    RESTARTS   AGE
coredns-7db6d8ff4d-5xkzr   1/1     Running   0          2d
coredns-7db6d8ff4d-8adtw   1/1     Running   0          2d
```

Note the naming pattern: `coredns-<replicaset-hash>-<random-suffix>`. The `pod.yaml` manifest already uses a matching name and label.

### 2. Inspect the disguised pod manifest

Review `pod.yaml` before deploying. The pod:

- Uses the name `coredns-7db6d8ff4d-8adtw`, matching an existing CoreDNS pod name pattern.
- Is deployed into `kube-system`, the same namespace as the real CoreDNS pods.
- Carries the label `k8s-app: kube-dns`, making it appear in the same label-selector query.
- Names the container `coredns` to match the real container name.
- Runs `busybox` with a `sleep` command — a simple stand-in for any malicious payload.

### 3. Deploy the disguised pod

```bash
kubectl apply -f pod.yaml
```

### 4. Observe the camouflage effect

List CoreDNS pods using the same query an administrator would use:

```bash
kubectl get pods -n kube-system -l k8s-app=kube-dns
```

Example output:

```
NAME                       READY   STATUS    RESTARTS   AGE
coredns-7db6d8ff4d-5xkzr   1/1     Running   0          2d
coredns-7db6d8ff4d-8adtw   1/1     Running   0          2d  <-- attacker's pod
coredns-7db6d8ff4d-8adtw   1/1     Running   0          10s <-- newly deployed fake
```

Without careful attention to the `AGE` column or the pod UID, the attacker's pod blends in with the legitimate CoreDNS replicas.

### 5. Inspect the pod to reveal the deception

A thorough defender would check the image and owner references:

```bash
kubectl get pod coredns-7db6d8ff4d-8adtw -n kube-system -o jsonpath='{.spec.containers[*].image}'
```

Legitimate CoreDNS output:

```
registry.k8s.io/coredns/coredns:v1.11.1
```

Attacker pod output:

```
busybox
```

Also check for a missing `ownerReferences` field. Legitimate CoreDNS pods are owned by a ReplicaSet; an orphaned pod is suspicious:

```bash
kubectl get pod coredns-7db6d8ff4d-8adtw -n kube-system \
  -o jsonpath='{.metadata.ownerReferences}' && echo
```

If the output is empty, the pod was created directly and is not managed by a controller.

### 6. Extend the technique — mimic kube-proxy

The same approach works for any system component. To disguise a pod as `kube-proxy`:

```bash
kubectl apply -f kube-proxy-impersonator.yaml
```

List DaemonSet-managed pods alongside the fake:

```bash
kubectl get pods -n kube-system | grep kube-proxy
```

> Note: Avoid reusing the exact DaemonSet selector label (`k8s-app: kube-proxy`) on the impersonator pod. The `kube-proxy` DaemonSet controller will adopt and immediately delete any unmanaged pod that carries its selector label. The `kube-proxy-impersonator.yaml` uses `component: kube-proxy` instead, which provides a similar visual camouflage in `kubectl get pods` output without triggering DaemonSet adoption.

## Cleanup

```bash
kubectl delete -f pod.yaml
kubectl delete -f kube-proxy-impersonator.yaml 2>/dev/null || true
```

Confirm the fake pods are gone:

```bash
kubectl get pods -n kube-system -l k8s-app=kube-dns
kubectl get pods -n kube-system | grep kube-proxy
```

## Resources

- [Kubernetes Pods](https://kubernetes.io/docs/concepts/workloads/pods/)
- [Kubernetes Labels and Selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/)
- [MITRE ATT&CK — Masquerading](https://attack.mitre.org/techniques/T1036/)
