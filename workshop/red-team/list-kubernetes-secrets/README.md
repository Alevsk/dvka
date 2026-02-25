# List Kubernetes Secrets

This document describes how an attacker can list Kubernetes secrets to gain access to sensitive information.

## Description

Kubernetes Secrets are objects for storing sensitive data, such as passwords and tokens. Attackers who have permissions to list secrets in the cluster can obtain the credentials that are stored in the secrets.

## Resources

- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)