# Access cloud resources

If the Kubernetes cluster is deployed in the cloud, in some cases attackers can leverage their access to a single container to get access to other cloud resources outside the cluster. For example, AKS uses several managed identities that are attached to the nodes, for the cluster operation. Similar identities exist also in EKS and GKE (EC2 roles and IAM service accounts, respectively). By default, running pods can retrieve the identities which in some configurations have privileged permissions. Therefore, if attackers gain access to a running pod in the cluster, they can leverage the identities to access external cloud resources.

Also, AKS has an option to authenticate with Azure using a service principal. When this option is enabled, each node stores service principal credentials that are located in `/etc/kubernetes/azure.json`. AKS uses this service principal to create and manage Azure resources that are needed for the cluster operation.

By default, the service principal has contributor permissions in the cluster’s Resource Group. Attackers who get access to this service principal file (by hostPath mount, for example) can use its credentials to access or modify the cloud resources.

## Resources

- <https://learn.microsoft.com/en-us/azure/aks/kubernetes-service-principal>
- <https://www.netspi.com/blog/technical/cloud-penetration-testing/extract-credentials-from-azure-kubernetes-service/>
