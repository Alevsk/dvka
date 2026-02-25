# Compromised Image in Registry

This document describes how an attacker can use a compromised image to gain access to the cluster.

## Description

Running a compromised image in a cluster can compromise the cluster. Attackers who get access to a private registry can plant their own compromised images in the registry. The latter can then be pulled by a user. In addition, users often use untrusted images from public registries (such as Docker Hub) that may be malicious.

## Resources

- [Supply Chain Threats Using Container Images](https://blog.aquasec.com/supply-chain-threats-using-container-images)
- [Malicious Docker Hub Container Images Cryptojacking](https://www.trendmicro.com/vinfo/us/security/news/virtualization-and-cloud/malicious-docker-hub-container-images-cryptocurrency-mining)