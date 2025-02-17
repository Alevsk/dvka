# Create New Kubernetes Cluster Using Kind

## Prerequisites

- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)

## Quick Start

1. Look at the cluster configuration in the `workshop-cluster.yaml` file

2. Create the cluster using the `kind` command

    ```bash
    kind create cluster --config workshop-cluster.yaml --name workshop-cluster
    ```

## Recommended

If using `kind`, once your kubernetes `workshop-cluster` is up and running you can push all the images in your local registry to the cluster

```bash
# push all images to you kind cluster
for image in $(cat ../../images.txt); do kind load docker-image $image --name workshop-cluster; done;
```

## Resources

- <https://kind.sigs.k8s.io/docs/user/configuration/>
