# Exec into Container

This document describes how an attacker can use `kubectl exec` to run commands in a container.

## Description

Attackers who have permissions, can run malicious commands in containers in the cluster using exec command (“kubectl exec”). In this method, attackers can use legitimate images, such as an OS image (e.g., Ubuntu) as a backdoor container, and run their malicious code remotely by using “kubectl exec”.

## Quick Start

1.  **Deploy a Pod**

    Deploy an NGINX pod to your cluster:

    ```bash
    kubectl apply -f nginx.yaml
    ```

2.  **Exec into the Pod**

    Exec into the pod and get a shell:

    ```bash
    kubectl exec -it nginx -- sh
    ```

3.  **Explore the Container**

    Once you have a shell to the container, you can explore its filesystem and run commands.

## Cleanup

    ```bash
    kubectl delete -f nginx.yaml
    ```

## Resources

- [kubectl exec](https://kubernetes.io/docs/reference/generated/kubectl/kubectl-commands#exec)