# Use Ubuntu 22.04 as base image
FROM mcr.microsoft.com/devcontainers/base:jammy

# Install basic system utilities and development tools
RUN apt-get update && apt-get install -y \
  git \
  python3 \
  python-is-python3 \
  python3-pip \
  make \
  ca-certificates \
  sudo \
  wget \
  unzip \
  vim \
  htop \
  net-tools \
  iputils-ping \
  tar \
  gzip \
  jq \
  apt-transport-https \
  gnupg \
  lsb-release \
  software-properties-common \
  curl \
  && rm -rf /var/lib/apt/lists/*

# Install Kind (Kubernetes in Docker)
RUN curl -Lo /usr/local/bin/kind https://kind.sigs.k8s.io/dl/v0.27.0/kind-linux-amd64 && \
  chmod +x /usr/local/bin/kind

# Install kubectl (Kubernetes command-line tool)
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
  chmod +x kubectl && \
  mv kubectl /usr/local/bin/

# Install Kustomize (Kubernetes configuration customization tool)
RUN curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash && \
  mv kustomize /usr/local/bin/

# Install K9s (Kubernetes CLI Terminal UI)
RUN curl -LO https://github.com/derailed/k9s/releases/download/v0.40.5/k9s_Linux_amd64.tar.gz && \
  tar -xzvf k9s_Linux_amd64.tar.gz && \
  mv k9s /usr/local/bin/ && \
  rm k9s_Linux_amd64.tar.gz

# Install kubectx (tool for switching between Kubernetes contexts)
RUN curl -LO https://raw.githubusercontent.com/ahmetb/kubectx/master/kubectx && \
  chmod +x kubectx && \
  mv kubectx /usr/local/bin/kubectx

# Install kubens (tool for switching between Kubernetes namespaces)
RUN curl -LO https://raw.githubusercontent.com/ahmetb/kubectx/master/kubens && \
  chmod +x kubens && \
  mv kubens /usr/local/bin/kubens


# Install mkcert (locally trusted certificates)
RUN curl -Lo /usr/local/bin/mkcert https://dl.filippo.io/mkcert/latest?for=linux/amd64 && \
  chmod +x /usr/local/bin/mkcert && \
  mkcert -install

# Install kube-hunter (Kubernetes penetration testing tool)
RUN pip install kube-hunter

# Install kube-linter (Kubernetes YAML linter)
RUN curl -Lo kube-linter.tar.gz https://github.com/stackrox/kube-linter/releases/download/v0.7.2/kube-linter-linux.tar.gz && \
  tar -zxvf kube-linter.tar.gz && \
  mv kube-linter /usr/local/bin/ && \
  chmod +x /usr/local/bin/kube-linter && \
  rm kube-linter.tar.gz

# Install terrascan (security vulnerability scanner)
RUN curl -Lo terrascan.tar.gz https://github.com/tenable/terrascan/releases/download/v1.19.9/terrascan_1.19.9_Linux_x86_64.tar.gz && \
  tar -zxvf terrascan.tar.gz && \
  mv terrascan /usr/local/bin/ && \
  chmod +x /usr/local/bin/terrascan && \
  rm terrascan.tar.gz

# Install kubeaudit (Kubernetes security auditing tool)
RUN curl -Lo kubeaudit.tar.gz https://github.com/Shopify/kubeaudit/releases/download/v0.22.2/kubeaudit_0.22.2_linux_amd64.tar.gz && \
  tar -zxvf kubeaudit.tar.gz && \
  mv kubeaudit /usr/local/bin/ && \
  chmod +x /usr/local/bin/kubeaudit && \
  rm kubeaudit.tar.gz

# Install nuclei (vulnerability scanner)
RUN curl -Lo nuclei_3.3.9_linux_amd64.zip https://github.com/projectdiscovery/nuclei/releases/download/v3.3.9/nuclei_3.3.9_linux_amd64.zip && \
  unzip -o nuclei_3.3.9_linux_amd64.zip && \
  mv nuclei /usr/local/bin/ && \
  chmod +x /usr/local/bin/nuclei && \
  rm nuclei_3.3.9_linux_amd64.zip

