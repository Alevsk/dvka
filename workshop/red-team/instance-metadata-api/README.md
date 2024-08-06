# Instance Metadata API

Cloud providers provide instance metadata service for retrieving information about the virtual machine, such as network configuration, disks, and SSH public keys. This service is accessible to the VMs via a non-routable IP address that can be accessed from within the VM only. Attackers who gain access to a container, may query the metadata API service for getting information about the underlying node.

For example, in Azure, the following request would retrieve all the metadata information of an instance: `http:///metadata/instance?api-version=2019-06-01`

## Resources
