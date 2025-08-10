# 🛠️ Kubernetes Security Workshop: Hands-On Attack and Defense

Learn everything you need to know to be proficient at `Kubernetes` security.

## Slides

- Talk - [Kubernetes Security: Attacking And Defending Modern Infrastructure](./resources/Kubernetes%20Security_%20Attacking%20And%20Defending%20Modern%20Infrastructure.pdf)
- Workshop - [Kubernetes Security: Hands-On Attack and Defense](https://docs.google.com/presentation/d/1algz44C9d2YFCHO3epScPzaEChJqBY2JijIjTtBnt7c/)

## Requirements

- [⚙️ Tools Used During The Workshop](./requirements.md)

## Lab Series

### 👶 Beginner

- [Create New Kubernetes Cluster Using Kind](./labs/create-cluster/README.md)
- [Explore ~/.kube/config File And Kubectl Command](./labs/explore-kubeconfig/README.md)
- [Explore k9s To Manage Your Cluster](./labs/explore-k9s/README.md)
- [Deploy Kubernetes Workload](./labs/deploy-workload/README.md)
- [Get a Shell to a Running Container](./labs/get-shell/README.md)
- [ConfigMaps & Secrets](./labs/configmaps-secrets/README.md)
- [Namespaces](./labs/namespaces/README.md)
- [Pod Security Context](./labs/pod-security-context/README.md)

### 👩‍💻 Intermediate

- [Kubernetes certificate authority](./labs/k8s-cert-authority/README.md)
- [cert-manager: X.509 certificate management for Kubernetes](./labs/cert-manager/README.md)
- [Pod Resource Limits](./labs/resource-limits/README.md)
- [Scratch Containers](./labs/scratch-containers/README.md)
- [Service Account Token](./labs/service-account-token/README.md)
- [Network Security Policies With Calico](./labs/network-policies-calico/README.md)

### 🥷 Advanced

- [Privilege Escalation Using Docker Containers](./labs/docker-privilege-escalation/README.md)
- [kube-bench: CIS Kubernetes Benchmark](./labs/kube-bench/README.md)
- [kube-hunter: Hunt for security weaknesses in Kubernetes clusters](./labs/kube-hunter/README.md)
- [kube-linter: Check Kubernetes YAML files and Helm charts](./labs/kube-linter/README.md)
- [terrascan: Static code analyzer for Infrastructure as Code](./labs/terrascan/README.md)
- [kubeaudit: Audit your Kubernetes clusters against common security controls](./labs/kubeaudit/README.md)
- [Deploy privileged container, A Container That Doesn't Contain Anything](./labs/privileged-container/README.md)
- [CVE-2025-1974 - Ingress Nightmare](./labs/ingress-nightmare/README.md)

## 🔴 Red Team

- [Common Attack Techniques](./red-team.md)
- [Attacking Kubernetes: Tools and tactics to compromise your first cluster](https://labs.iximiuz.com/trainings/kubernetes-security-workshop-hands-on-attack-and-defense-e925c0d3)

## 🟡 Blue Team

- Digital forensics
- telemtry
- gitops
- hardening
