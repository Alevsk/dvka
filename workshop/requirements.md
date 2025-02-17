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
apt autoremove completed .... ✅
```

## Container Images

The [images.txt](./images.txt) file has a list of all the container images used during this workshop. To save time, download all of them and save them into your local registry.

```bash
# pull all images to your local registry
for image in $(cat images.txt); do docker pull $image; done;
```

### (OPTIONAL) Load container images from a backup file

Use the `workshop-images.sh` script to load all the container images from a provided backup file. Download the `docker_images.zip` file using the following [link 🔗 (size: 2.1G)](https://drive.google.com/file/d/1wM9sW-AdZibeGnR4058uCXQZwmguoQd_/view).

> *If you see a rate-limit message make sure you are logged in with your google/gmail account before trying to download the file*

```bash
chmod +x workshop-images.sh
./workshop-images.sh load --file docker_images.zip
```

### Verify Images

Verify all the images are in your local registry

```bash
docker images
```

## Virtual Machine

Alternatively, the following **virtual machine** image contains everything you need to go over the labs:

1. Install the right version of [Oracle VM VirtualBox](https://www.virtualbox.org/wiki/Downloads) or [VMware Workstation Player](https://www.vmware.com/products/workstation-player/workstation-player-evaluation.html) for your system.

1. Download the `Kubernetes Security Workshop` image using the following [link 🔗](https://drive.google.com/file/d/12IX4xGvfqgZLrtutimWqQdxpJRRzDPto/view) (size: 26.5G).

   > *If you see a rate-limit message make sure you are logged in with your google/gmail account before trying to download the file*

1. Run VirtualBox / VMware Player and import the virtual machine image (virtual machine size: 100G once imported).

1. Login into the virtual machine.

> 🔒 Credentials - **username:** kubernetes / **password:** security

![virtual machine](./images/virtual-machine.jpeg)
