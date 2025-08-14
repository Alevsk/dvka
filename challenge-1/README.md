# Hack The NFT Museum – Challenge 1

Welcome to the NFT Museum challenge! In this scenario, you'll explore common security vulnerabilities in cloud-native environments by attempting to compromise a modern NFT marketplace running on Kubernetes.

## Scenario Overview

You've discovered a new NFT marketplace that claims to be revolutionizing the digital art world. However, your security research suggests that the application might have some critical vulnerabilities. Your mission is to investigate the platform and find potential security flaws that could lead to a compromise of the application.

![NFT Museum Store](./docs/images/nft-store.jpg)

## Challenge Flags

Your objective is to find the hidden flag by exploiting a series of vulnerabilities in the application. The path to success involves:

* **Initial Access**
  * Start by exploring the NFT marketplace's web interface
  * Look for ways to manipulate the application's file handling mechanisms
  * The flag is hidden somewhere in the infrastructure, but finding it will require chaining multiple vulnerabilities

## Attack Chain Overview

This challenge simulates a realistic attack path that combines multiple security concepts:

1. **Web Application Exploitation**: Begin by identifying and exploiting web application vulnerabilities
2. **Infrastructure Interaction**: Learn how the application interacts with its underlying Kubernetes infrastructure
3. **Kubernetes API Access**: Understand how to leverage compromised credentials to interact with the Kubernetes API

## Key Security Concepts

* **Directory Traversal**
  * Understanding how applications handle file paths
  * Exploiting improper file path validation

* **Server-Side Request Forgery (SSRF)**
  * How internal services can be accessed through vulnerable applications
  * The implications of SSRF in cloud-native environments

* **Kubernetes Security**
  * Service account tokens and their capabilities
  * Interaction with the Kubernetes API server
  * Container escape techniques

For technical setup and deployment instructions, please refer to the [Development Guide](./code/README.md) file.

Good luck, and happy hacking!
