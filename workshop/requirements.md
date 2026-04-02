# ⚙️ Tools Used During The Workshop

The following tools are required during the workshop

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
- [nuclei](https://github.com/projectdiscovery/nuclei/releases/download/v3.2.9/nuclei_3.2.9_linux_amd64.zip)
- [trivy](https://github.com/aquasecurity/trivy)
- [skopeo](https://github.com/containers/skopeo)

> [!IMPORTANT]  
> There are several ways to start a virtual machine with all the required tools for this workshop, you have to choose one of the following options

## labs.iximiuz.com (beta)

I'm experimenting with [labs.iximiuz.com](https://labs.iximiuz.com/) to host the Kubernetes Security Workshop training material.

1. Sign up for [labs.iximiuz.com](https://labs.iximiuz.com/).
2. Join the workshop at <https://labs.iximiuz.com/trainings/kubernetes-security-workshop-hands-on-attack-and-defense-e925c0d3>.

## Install Script

You can install the required tools for this workshop by running the `install-tools.sh` script, a **Virtual Machine** with [Kali](https://www.kali.org/get-kali/#kali-installer-images) (Recommended), [Ubuntu](https://ubuntu.com/download/desktop) or [Debian](https://www.debian.org/distrib/) is recommended to run the script.

```bash
# make it executable
chmod +x install-tools.sh
# install all required tools for this workshop 
sudo ./install-tools.sh --install
```

Example output:

```bash
Package lists updated .... ✅ 
Prerequisites installed .... ✅ 
Docker installed .... ✅ 
Kind installed .... ✅ 
kubectl installed .... ✅ 
Kustomize installed .... ✅ 
k9s installed .... ✅ 
mkcert installed .... ✅ 
kube-hunter installed .... ✅ 
kube-linter installed .... ✅ 
terrascan installed .... ✅ 
kubeaudit installed .... ✅ 
nuclei installed .... ✅
trivy installed .... ✅
skopeo installed .... ✅
apt autoremove completed .... ✅
```
