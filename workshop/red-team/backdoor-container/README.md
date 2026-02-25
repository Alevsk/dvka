# Backdoor Container

This document describes how an attacker can use a backdoor container to maintain persistence in the cluster.

## Description

Attackers run their malicious code in a container in the cluster. By using the Kubernetes controllers such as DaemonSets or Deployments, attackers can ensure that a constant number of containers run in one, or all, the nodes in the cluster.

## Resources

- [Kubernetes DaemonSets](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)
- [Kubernetes Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)