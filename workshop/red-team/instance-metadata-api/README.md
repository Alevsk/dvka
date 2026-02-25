# Instance Metadata API

This document describes how an attacker can use the instance metadata API to gain information about the underlying node.

## Description

Cloud providers provide instance metadata service for retrieving information about the virtual machine, such as network configuration, disks, and SSH public keys. This service is accessible to the VMs via a non-routable IP address that can be accessed from within the VM only. Attackers who gain access to a container, may query the metadata API service for getting information about the underlying node.

For example, in Azure, the following request would retrieve all the metadata information of an instance: `http:///metadata/instance?api-version=2019-06-01`

## Resources

- [Azure Instance Metadata Service](https://learn.microsoft.com/en-us/azure/virtual-machines/windows/instance-metadata-service)
- [GCP Instance Metadata](https://cloud.google.com/compute/docs/storing-retrieving-metadata)
- [AWS Instance Metadata](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/instancedata-data-retrieval.html)