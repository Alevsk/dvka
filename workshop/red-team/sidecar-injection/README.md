# Sidecar injection

A Kubernetes Pod is a group of one or more containers with shared storage and network resources. Sidecar container is a term that is used to describe an additional container that resides alongside the main container. For example, service-mesh proxies are operating as sidecars in the applications’ pods. Attackers can run their code and hide their activity by injecting a sidecar container to a legitimate pod in the cluster instead of running their own separated pod in the cluster.

## Resources
