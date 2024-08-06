# Access Managed Identity credentials

Managed identities are identities that are managed by the cloud provider and can be allocated to cloud resources, such as virtual machines. Those identities are used to authenticate with cloud services. The identity’s secret is fully managed by the cloud provider, which eliminates the need to manage the credentials. Applications can obtain the identity’s token by accessing the Instance Metadata Service (IMDS). Attackers who get access to a Kubernetes pod can leverage their access to the IMDS endpoint to get the managed identity’s token. With a token, the attackers can access cloud resources.

## References
