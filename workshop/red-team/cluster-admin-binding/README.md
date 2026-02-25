# Cluster-Admin Binding

This document describes how an attacker can create a cluster-admin binding to gain administrative privileges in the cluster.

## Description

Role-based access control (RBAC) is a key security feature in Kubernetes. RBAC can restrict the allowed actions of the various identities in the cluster. Cluster-admin is a built-in high privileged role in Kubernetes. Attackers who have permissions to create bindings and cluster-bindings in the cluster can create a binding to the cluster-admin ClusterRole or to other high privileges roles.

## Resources

- [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)