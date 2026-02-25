# Getting a Shell to a Running Container

## Prerequisites

- A running Kubernetes cluster.
- `kubectl` installed and configured to connect to your cluster.

## Quick Start

1. Deploy busybox as a pod (notice we are not creating a `deployment` this time)

    ```bash
    # create busybox pod
    kubectl apply -f busybox.yaml
    ```

2. Exec into the running container

    **kubectl:**

    ```bash
    kubectl exec -it pod/busybox -- sh
    ```

    **k9s:**

    Pods `>` busybox `>` press `<s>`

3. Explore the container file system

    - `top` command
    - `ls` (/proc, /sys, /dev, /etc) command
    - `printenv`

4. Terminate busybox pod

    ```bash
    kubectl delete -f busybox.yaml
    ```

## Resources

- [Get a Shell to a Running Container](https://kubernetes.io/docs/tasks/debug/debug-application/get-shell-running-container/)