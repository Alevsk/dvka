# Malicious Admission Controller

This document describes how an attacker can use a malicious admission controller to gain persistence in the cluster.

## Description

Admission controllers are a powerful Kubernetes feature that can be used to enforce security policies in the cluster. Admission controllers intercept and process requests to the Kubernetes API server. There are two types of admission controllers: validating admission controllers and mutating admission controllers. Validating admission controllers can only approve or deny requests, while mutating admission controllers can also modify the requests.

Attackers who have permissions to create and modify admission controllers can use them for various malicious purposes. For example, attackers can use a mutating admission controller to inject their malicious code into any new pod that is created in the cluster.

## Resources

- [Kubernetes Admission Controllers](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/)