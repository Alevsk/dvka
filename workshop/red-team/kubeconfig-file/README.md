# Kubeconfig File

This document describes how an attacker can use a kubeconfig file to gain access to the cluster.

## Description

The kubeconfig file, also used by kubectl, contains details about Kubernetes clusters including their location and credentials. If the cluster is hosted as a cloud service (such as AKS or GKE), this file is downloaded to the client via cloud commands (e.g., `az aks get-credential` for AKS or `gcloud container clusters get-credentials` for GKE).

If attackers get access to this file, for instance via a compromised client, they can use it for accessing the clusters.

## Quick Start

1.  **View the Kubeconfig File**

    View the contents of your kubeconfig file:

    ```bash
    cat ~/.kube/config
    ```

## Resources

- [Organizing Cluster Access Using kubeconfig Files](https://kubernetes.io/docs/concepts/configuration/organize-cluster-access-kubeconfig/)