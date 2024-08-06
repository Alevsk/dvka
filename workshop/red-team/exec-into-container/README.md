# Exec into container

Attackers who have permissions, can run malicious commands in containers in the cluster using exec command (“kubectl exec”). In this method, attackers can use legitimate images, such as an OS image (e.g., Ubuntu) as a backdoor container, and run their malicious code remotely by using “kubectl exec”.

## Quick Start

1. Deploy nginx as a pod (notice we are not creating a `deployment` this time)

    ```bash
    # create nginx pod
    kubectl apply -f nginx.yaml
    ```

2. Exec into the running container

    **kubectl:**

    ```bash
    kubectl exec -it pod/nginx -- sh
    ```

3. Explore the nginx container file system

    - `top` command
    - `ls` (/proc, /sys, /dev, /etc) command
    - `printenv`

4. Terminate nginx pod

    ```bash
    kubectl delete -f nginx.yaml
    ```

## Resouces

- <https://kubernetes.io/docs/tasks/debug/debug-application/get-shell-running-container/>
- <https://kubernetes-threat-matrix.redguard.ch/execution/bash-cmd-in-container/>
