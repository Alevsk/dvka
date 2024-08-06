# Application credentials in configuration files

Developers store secrets in the Kubernetes configuration files, such as environment variables in the pod configuration. Such behavior is commonly seen in clusters that are monitored by Microsoft Defender for Cloud. Attackers who have access to those configurations, by querying the API server or by accessing those files on the developer’s endpoint, can steal the stored secrets and use them.

Using those credentials attackers may gain access to additional resources inside and outside the cluster.

## References
