# CoreDNS poisoning

CoreDNS is a modular Domain Name System (DNS) server written in Go, hosted by Cloud Native Computing Foundation (CNCF). CoreDNS is the main DNS service that is being used in Kubernetes. The configuration of CoreDNS can be modified by a file named corefile. In Kubernetes, this file is stored in a ConfigMap object, located at the kube-system namespace. If attackers have permissions to modify the ConfigMap, for example by using the container’s service account, they can change the behavior of the cluster’s DNS, poison it, and take the network identity of other services.

## Resources

- <https://coredns.io/>
