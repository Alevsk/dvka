# Mount Service Principal

This document describes how an attacker can mount a service principal to gain access to cloud resources.

## Description

AKS has an option to authenticate with Azure using a service principal. When this option is enabled, each node stores service principal credentials that are located in `/etc/kubernetes/azure.json`. AKS uses this service principal to create and manage Azure resources that are needed for the cluster operation. By default, the service principal has contributor permissions in the cluster’s Resource Group. Attackers who get access to this service principal file (by hostPath mount, for example) can use its credentials to access or modify the cloud resources.

## Resources

- [AKS Service Principals](https://learn.microsoft.com/en-us/azure/aks/kubernetes-service-principal)
- [Extracting Credentials from Azure Kubernetes Service](https://www.netspi.com/blog/technical/cloud-penetration-testing/extract-credentials-from-azure-kubernetes-service/)
