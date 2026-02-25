# Kubernetes Security Workshop

This workshop provides a hands-on introduction to Kubernetes security, covering both attack and defense.

## Slides

- [Kubernetes Security: Attacking and Defending Modern Infrastructure](./resources/Kubernetes%20Security_%20Attacking%20And%20Defending%20Modern%20Infrastructure.pdf)
- [Kubernetes Security: Hands-On Attack and Defense](https://docs.google.com/presentation/d/1algz44C9d2YFCHO3epScPzaEChJqBY2JijIjTtBnt7c/)

## Requirements

- [Tools](./requirements.md)

## Labs

### Beginner

- [Creating a Kubernetes Cluster with Kind](./labs/create-cluster/README.md)
- [Exploring the `kubeconfig` File and `kubectl`](./labs/explore-kubeconfig/README.md)
- [Exploring Your Cluster with k9s](./labs/explore-k9s/README.md)
- [Deploying a Kubernetes Workload](./labs/deploy-workload/README.md)
- [Getting a Shell to a Running Container](./labs/get-shell/README.md)
- [Managing Configuration with ConfigMaps and Secrets](./labs/configmaps-secrets/README.md)
- [Working with Namespaces](./labs/namespaces/README.md)
- [Pod Security Context](./labs/pod-security-context/README.md)

### Intermediate

- [Kubernetes Certificate Authority](./labs/k8s-cert-authority/README.md)
- [cert-manager: X.509 Certificate Management for Kubernetes](./labs/cert-manager/README.md)
- [Pod Resource Limits](./labs/resource-limits/README.md)
- [Scratch Containers](./labs/scratch-containers/README.md)
- [Service Account Tokens](./labs/service-account-token/README.md)
- [Network Policies with Calico](./labs/network-policies-calico/README.md)

### Advanced

- [Privilege Escalation with Docker](./labs/docker-privilege-escalation/README.md)
- [Kube-bench: CIS Kubernetes Benchmark](./labs/kube-bench/README.md)
- [kube-hunter: Hunt for Security Weaknesses in Kubernetes Clusters](./labs/kube-hunter/README.md)
- [KubeLinter: Static Analysis for Kubernetes YAML Files and Helm Charts](./labs/kube-linter/README.md)
- [Terrascan: Static Code Analysis for Infrastructure as Code](./labs/terrascan/README.md)
- [kubeaudit: Audit Your Kubernetes Clusters](./labs/kubeaudit/README.md)
- [Privileged Containers](./labs/privileged-container/README.md)
- [CVE-2021-25742: Ingress-NGINX Annotation Injection](./labs/ingress-nightmare/README.md)

## Red Team

- [Red Team Techniques](./red-team.md)

## Blue Team

- [Blue Team Techniques](./blue-team.md)