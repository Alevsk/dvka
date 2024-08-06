# Network mapping

Attackers may try to map the cluster network to get information on the running applications, including scanning for known vulnerabilities. By default, there is no restriction on pods communication in Kubernetes. Therefore, attackers who gain access to a single container, may use it to probe the network.

## Resources

- <https://github.com/projectdiscovery/naabu>
