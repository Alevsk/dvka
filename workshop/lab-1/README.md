# Create New Kubernetes Cluster Using Kind

## Prerequisites

- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)

## Quick Start

1. Look at the cluster configuration in the `workshop-cluster.yaml` file

    ```yaml
    kind: Cluster
    apiVersion: kind.x-k8s.io/v1alpha4
    networking:
      apiServerAddress: "0.0.0.0"
      apiServerPort: 6443
    nodes:
      - role: control-plane
        image: kindest/node:v1.27.3
      - role: worker
        image: kindest/node:v1.27.3
      - role: worker
        image: kindest/node:v1.27.3
      - role: worker
        image: kindest/node:v1.27.3
    ```

2. Create the cluster

    ```bash
    kind create cluster --config workshop-cluster.yaml --name workshop-cluster
    ```

## Resouces

- <https://kind.sigs.k8s.io/docs/user/configuration/>
