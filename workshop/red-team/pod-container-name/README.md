# Pod / Container Name Similarity

This document describes how an attacker can use pod or container name similarity to hide their malicious activity.

## Description

Attackers may give their pods and containers names that are similar to the names of other objects in the cluster. This can be used to hide their malicious activity from the cluster administrator.

## Quick Start

1. List `coredns` pods running in the cluster

    ```bash
    kubectl get pods -n kube-system -l k8s-app=kube-dns
    ```

    Observe how the pods have the following format `coredns-{random suffix}` in their names.

    Edit `pod.yaml` file. Set a name using a similar suffix, ie: `coredns-7db6d8ff4d-8adtw`

2. Create the new busybox pod by running the following command

    ```bash
    kubectl apply -f pod.yaml
    ```

3. List all the `coredns` pods again

    ```bash
    kubectl get pods -n kube-system -l k8s-app=kube-dns
    ```

    Your new `coredns` pod should be there, this pod in reality is running `busybox`

4. Terminate the pod

    ```bash
    kubectl delete -f pod.yaml
    ```

## Resources

- [Kubernetes Pods](https://kubernetes.io/docs/concepts/workloads/pods/)