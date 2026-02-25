# Using Cloud Credentials

This document describes how an attacker can use compromised cloud credentials to gain access to the cluster.

## Description

If attackers get access to cloud credentials, they can use them to access the cluster. For example, in AKS, users can be authenticated with Azure Active Directory (Azure AD). Attackers who get access to the credentials of a user with permissions to the cluster can use them to access the cluster.

## Resources

- [AKS - Azure AD Integration](https://learn.microsoft.com/en-us/azure/aks/managed-aad)