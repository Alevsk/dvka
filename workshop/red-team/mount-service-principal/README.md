# Mount service principal

When the cluster is deployed in the cloud, in some cases attackers can leverage their access to a container in the cluster to gain cloud credentials. For example, in AKS each node contains service principal credential.

## Resources

- <https://learn.microsoft.com/en-us/azure/aks/kubernetes-service-principal>
- <https://www.netspi.com/blog/technical/cloud-penetration-testing/extract-credentials-from-azure-kubernetes-service/>