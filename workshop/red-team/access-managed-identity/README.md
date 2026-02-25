# Accessing Managed Identity Credentials

This document describes how an attacker can access managed identity credentials to gain access to cloud resources.

## Description

Managed identities are identities that are managed by the cloud provider and can be allocated to cloud resources, such as virtual machines. Those identities are used to authenticate with cloud services. The identity’s secret is fully managed by the cloud provider, which eliminates the need to manage the credentials. Applications can obtain the identity’s token by accessing the Instance Metadata Service (IMDS). Attackers who get access to a Kubernetes pod can leverage their access to the IMDS endpoint to get the managed identity’s token. With a token, the attackers can access cloud resources.

## Resources

- [Azure Managed Identities](https://learn.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview)
- [GCP Managed Identities](https://cloud.google.com/iam/docs/managed-identities)
- [AWS IAM Roles for Service Accounts](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)