# Writable hostPath Mount

This document describes how an attacker can use a writable hostPath mount to escape to the host.

## Description

A hostPath volume mounts a file or directory from the host node’s filesystem into your Pod. Attackers who have permissions to create pods with a writable hostPath mount can use it to escape the container and get access to the host.

## Resources

- [Kubernetes hostPath](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath)