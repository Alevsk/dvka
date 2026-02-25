# Bash/cmd inside Container

This document describes how an attacker can use `kubectl exec` to run commands in a container.

## Description

Attackers who have permissions to run a cmd/bash script inside a container can use it to execute malicious code and compromise cluster resources.

## Quick Start

1.  **Run a Command in a Container**

    Run a command in a new container:

    ```bash
    kubectl run bash --restart=Never -ti --rm --image busybox -- nslookup google.com
    ```

2.  **Get a Shell to a Container**

    Get a shell to a new container:

    ```bash
    kubectl run nginx --restart=Never -ti --rm --image nginx -- bash
    ```

## Resources

- [kubectl exec](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#exec)