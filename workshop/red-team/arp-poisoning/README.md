# ARP poisoning and IP spoofing

Kubernetes has numerous network plugins (Container Network Interfaces or CNIs) that can be used in the cluster. Kubenet is the basic, and in many cases the default, network plugin. In this configuration, a bridge is created on each node (cbr0) to which the various pods are connected using veth pairs. The fact that cross-pod traffic is through a bridge, a level-2 component, means that performing ARP poisoning in the cluster is possible. Therefore, if attackers get access to a pod in the cluster, they can perform ARP poisoning, and spoof the traffic of other pods. By using this technique, attackers can perform several attacks at the network-level which can lead to lateral movements, such as DNS spoofing or stealing cloud identities of other pods (`CVE-2021-1677`).

## Refresh
