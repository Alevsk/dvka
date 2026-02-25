# Static Pods

This document describes how an attacker can use static pods to maintain persistence on a node.

## Description

Static pods are pods that are managed directly by the kubelet daemon on a specific node, without the API server observing them. Static pods are always bound to one Kubelet on a specific node. The kubelet automatically tries to create a mirror pod on the Kubernetes API server for each static pod. This means that the pods that are running on the nodes are visible on the API server, but cannot be controlled from there.

Attackers who have access to the host can create a static pod and run their malicious code in it.

## Resources

- [Static Pods](https://kubernetes.io/docs/tasks/configure-pod-container/static-pod/)