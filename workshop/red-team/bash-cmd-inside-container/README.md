# Bash/cmd inside container

Attackers who have permissions to run a cmd/bash script inside a container can use it to execute malicious code and compromise cluster resources.

## Quick Start

1. Run command inside the cluster using a container

    ```bash
    # run nslookup command
    kubectl run bash --restart=Never -ti --rm --image busybox -- nslookup google.com 
    # run bash command
    kubectl run nginx --restart=Never -ti --rm --image nginx -- bash
    ```

## Resources

- <https://kubernetes.io/docs/tasks/debug/debug-application/get-shell-running-container/>
