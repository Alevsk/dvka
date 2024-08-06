# Images from a private registry

The images that are running in the cluster can be stored in a private registry. For pulling those images, the container runtime engine (such as Docker or containerd) needs to have valid credentials to those registries. If the registry is hosted by the cloud provider, in services like Azure Container Registry (ACR) or Amazon Elastic Container Registry (ECR), cloud credentials are used to authenticate to the registry. If attackers get access to the cluster, in some cases they can obtain access to the private registry and pull its images. For example, attackers can use the managed identity token as described in the “Access managed identity credential” technique. Similarly, in EKS, attackers can use the AmazonEC2ContainerRegistryReadOnly policy that is bound by default to the node’s IAM role.

## Resources
