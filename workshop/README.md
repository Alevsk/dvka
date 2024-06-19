# 🛠️ Kubernetes Security Workshop: Hands-On Attack and Defense

Learn everything you need to know to be proficient at `Kubernetes` security.

## ⚙️ Tools Used During The Workshop

The provided **virtual machine** contains everything you need to go over the labs:

1. Install the right version of [Oracle VM VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [VMware Workstation Player](https://www.vmware.com/products/workstation-player/workstation-player-evaluation.html) for your system.

1. Download the `Kubernetes Security Workshop` image using the following [link 🔗](https://drive.google.com/file/d/12IX4xGvfqgZLrtutimWqQdxpJRRzDPto/view) (size: 26.5G).

1. Run VirtualBox / VMware Player and import the virtual machine image (virtual machine size: 100G once imported).

1. Login into the virtual machine.

> 🔒 Credentials - **username:** kubernetes / **password:** security

![virtual machine](./images/virtual-machine.jpeg)

---

Alternatively, you can manually install the following tools on your system (Linux & Mac OSX):

- [jq](https://jqlang.github.io/jq/)
- [Docker](https://docs.docker.com/engine/install/)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [Kustomize](https://kustomize.io/)
- [k9s](https://k9scli.io/topics/install/)
- [mkcert](https://github.com/FiloSottile/mkcert)
- [kube-bench](https://raw.githubusercontent.com/aquasecurity/kube-bench/main/job.yaml)
- [kube-hunter](https://github.com/aquasecurity/kube-hunter)
- [kube-linter](https://github.com/stackrox/kube-linter/releases/download/v0.6.5/kube-linter-linux.tar.gz)
- [terrascan](https://github.com/tenable/terrascan/releases/download/v1.18.3/terrascan_1.18.3_Linux_x86_64.tar.gz)
- [kubeaudit](https://github.com/Shopify/kubeaudit/releases/download/v0.22.0/kubeaudit_0.22.0_linux_amd64.tar.gz)

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
