#!/bin/bash

set -e
set -o pipefail

TOOLS_DIR="/usr/local/bin"
MKCERT_DIR="$HOME/.local/share/mkcert"
WORKSHOP_DIR="/tmp/workshop"
USER=$(logname)
DEBUG=0
INSTALL_TOOLS=()
INSTALL=0

# Ensure we are running as root
if [[ $EUID -ne 0 ]]; then
  printf "This script must be run as root. Use sudo.\n" >&2
  exit 1
fi

# Function to display help menu
show_help() {
  cat << EOF
Usage: ${0##*/} [--debug] [--install TOOL1,TOOL2,...] [--help]

Options:
  --debug                Enable debug mode
  --install TOOL1,TOOL2  Install the specified tools (comma-separated list)
  --help                 Display this help menu

Available tools:
  docker
  kind
  kubectl
  kustomize
  k9s
  mkcert
  kube_hunter
  kube_linter
  terrascan
  kubeaudit
  nuclei
EOF
}

# Check for flags
while [[ $# -gt 0 ]]; do
  case "$1" in
    --debug)
      DEBUG=1
      shift
      ;;
    --install)
      INSTALL=1
      if [[ -n "$2" && "$2" != --* ]]; then
        IFS=',' read -ra INSTALL_TOOLS <<< "$2"
        shift 2
      else
        INSTALL_TOOLS=()
        shift
      fi
      ;;
    --install=*)
      INSTALL=1
      IFS=',' read -ra INSTALL_TOOLS <<< "${1#*=}"
      shift
      ;;
    --help)
      show_help
      exit 0
      ;;
    *)
      printf "Unknown option: %s\n" "$1"
      show_help
      exit 1
      ;;
  esac
done

# Logging functions
log_info() {
  printf "\r%s .... ⏳" "$1"
}

log_done() {
  printf "\r%s .... ✅\n" "$1"
}

run_cmd() {
  if [[ $DEBUG -eq 1 ]]; then
    eval "$1"
  else
    eval "$1" &> /dev/null
  fi
}

# Create WORKSHOP_DIR if it doesn't exist
mkdir -p "$WORKSHOP_DIR"

# Function to update package lists
update_package_lists() {
  log_info "Updating package lists"
  run_cmd "apt-get update -y"
  log_done "Package lists updated"
}

# Function to install prerequisites
install_prerequisites() {
  log_info "Installing prerequisites"
  run_cmd "apt-get install -y ca-certificates curl wget tar gzip jq apt-transport-https gnupg lsb-release software-properties-common python3 python3-pip python-is-python3 unzip"
  log_done "Prerequisites installed"
}

# Function to install Docker
install_docker() {
  log_info "Installing Docker"
  # Uninstall previous conflicting packages
  for pkg in docker.io containerd runc; do
    run_cmd "apt-get remove -y $pkg"
  done

  if ! command -v docker &> /dev/null; then
    DISTRIBUTOR=$(lsb_release -is)
    case "$DISTRIBUTOR" in
      Kali)
        run_cmd "mkdir -p /etc/apt/keyrings"
        run_cmd "curl -fsSL https://download.docker.com/linux/debian/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg"
        run_cmd "echo 'deb [arch=amd64 signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/debian bookworm stable' | tee /etc/apt/sources.list.d/docker.list"
        run_cmd "apt-get update -y"
        run_cmd "apt-get install -y docker-ce docker-ce-cli containerd.io"
        ;;
      Debian)
        run_cmd "install -m 0755 -d /etc/apt/keyrings"
        run_cmd "curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc"
        run_cmd "chmod a+r /etc/apt/keyrings/docker.asc"
        run_cmd "echo 'deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian $(. /etc/os-release && echo $VERSION_CODENAME) stable' | tee /etc/apt/sources.list.d/docker.list > /dev/null"
        run_cmd "apt-get update -y"
        run_cmd "apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin"
        ;;
      Ubuntu|Pop)
        run_cmd "install -m 0755 -d /etc/apt/keyrings"
        run_cmd "curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc"
        run_cmd "chmod a+r /etc/apt/keyrings/docker.asc"
        run_cmd "echo 'deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo $VERSION_CODENAME) stable' | tee /etc/apt/sources.list.d/docker.list > /dev/null"
        run_cmd "apt-get update -y"
        run_cmd "apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin"
        ;;
      *)
        printf "Unsupported distribution: %s\n" "$DISTRIBUTOR" >&2
        exit 1
        ;;
    esac

    run_cmd "systemctl start docker"
    run_cmd "systemctl enable docker"
  fi
  # Create docker group and add user to it
  if ! getent group docker > /dev/null; then
    run_cmd "groupadd docker"
    run_cmd "usermod -aG docker $USER"
    run_cmd "newgrp docker"
  fi
  log_done "Docker installed"
}

# Function to install Kind
install_kind() {
  log_info "Installing Kind"
  if ! command -v kind &> /dev/null; then
    run_cmd "curl -Lo /usr/local/bin/kind https://kind.sigs.k8s.io/dl/v0.27.0/kind-linux-amd64"
    run_cmd "chmod +x /usr/local/bin/kind"
  fi
  log_done "Kind installed"
}

# Function to install kubectl
install_kubectl() {
  log_info "Installing kubectl"
  if ! command -v kubectl &> /dev/null; then
    run_cmd "curl -Lo /usr/local/bin/kubectl https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    run_cmd "chmod +x /usr/local/bin/kubectl"
  fi
  log_done "kubectl installed"
}

# Function to install Kustomize
install_kustomize() {
  log_info "Installing Kustomize"
  if ! command -v kustomize &> /dev/null; then
    run_cmd "curl -Lo $WORKSHOP_DIR/kustomize.tar.gz https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv5.6.0/kustomize_v5.6.0_linux_amd64.tar.gz"
    run_cmd "tar -zxvf $WORKSHOP_DIR/kustomize.tar.gz -C $WORKSHOP_DIR"
    run_cmd "mv $WORKSHOP_DIR/kustomize /usr/local/bin/"
    run_cmd "chmod +x /usr/local/bin/kustomize"
    run_cmd "rm $WORKSHOP_DIR/kustomize.tar.gz"
  fi
  log_done "Kustomize installed"
}

# Function to install k9s
install_k9s() {
  log_info "Installing k9s"
  if ! command -v k9s &> /dev/null; then
    run_cmd "curl -Lo $WORKSHOP_DIR/k9s.tar.gz https://github.com/derailed/k9s/releases/download/v0.40.5/k9s_Linux_amd64.tar.gz"
    run_cmd "tar -zxvf $WORKSHOP_DIR/k9s.tar.gz -C $WORKSHOP_DIR"
    run_cmd "mv $WORKSHOP_DIR/k9s /usr/local/bin/"
    run_cmd "chmod +x /usr/local/bin/k9s"
    run_cmd "rm $WORKSHOP_DIR/k9s.tar.gz $WORKSHOP_DIR/LICENSE $WORKSHOP_DIR/README.md"
  fi
  log_done "k9s installed"
}

# Function to install mkcert
install_mkcert() {
  log_info "Installing mkcert"
  if ! command -v mkcert &> /dev/null; then
    run_cmd "curl -Lo $WORKSHOP_DIR/mkcert https://dl.filippo.io/mkcert/latest?for=linux/amd64"
    run_cmd "chmod +x $WORKSHOP_DIR/mkcert"
    run_cmd "mv $WORKSHOP_DIR/mkcert /usr/local/bin/"
    run_cmd "mkcert -install"
  fi
  log_done "mkcert installed"
}

# Function to install kube-hunter
install_kube_hunter() {
  log_info "Installing kube-hunter"
  if ! command -v kube-hunter &> /dev/null; then
    run_cmd "curl -Lo $WORKSHOP_DIR/kube-hunter https://github.com/aquasecurity/kube-hunter/releases/download/v0.6.8/kube-hunter-linux-x86_64-refs.tags.v0.6.8"
    run_cmd "mv $WORKSHOP_DIR/kube-hunter /usr/local/bin/"
    run_cmd "chmod +x /usr/local/bin/kube-hunter"
  fi
  log_done "kube-hunter installed"
}

# Function to install kube-linter
install_kube_linter() {
  log_info "Installing kube-linter"
  if ! command -v kube-linter &> /dev/null; then
    run_cmd "curl -Lo $WORKSHOP_DIR/kube-linter.tar.gz https://github.com/stackrox/kube-linter/releases/download/v0.7.2/kube-linter-linux.tar.gz"
    run_cmd "tar -zxvf $WORKSHOP_DIR/kube-linter.tar.gz -C $WORKSHOP_DIR"
    run_cmd "mv $WORKSHOP_DIR/kube-linter /usr/local/bin/"
    run_cmd "chmod +x /usr/local/bin/kube-linter"
    run_cmd "rm $WORKSHOP_DIR/kube-linter.tar.gz $WORKSHOP_DIR/LICENSE $WORKSHOP_DIR/README.md"
  fi
  log_done "kube-linter installed"
}

# Function to install terrascan
install_terrascan() {
  log_info "Installing terrascan"
  if ! command -v terrascan &> /dev/null; then
    run_cmd "curl -Lo $WORKSHOP_DIR/terrascan.tar.gz https://github.com/tenable/terrascan/releases/download/v1.19.9/terrascan_1.19.9_Linux_x86_64.tar.gz"
    run_cmd "tar -zxvf $WORKSHOP_DIR/terrascan.tar.gz -C $WORKSHOP_DIR"
    run_cmd "mv $WORKSHOP_DIR/terrascan /usr/local/bin/"
    run_cmd "chmod +x /usr/local/bin/terrascan"
    run_cmd "rm $WORKSHOP_DIR/terrascan.tar.gz $WORKSHOP_DIR/CHANGELOG.md $WORKSHOP_DIR/LICENSE $WORKSHOP_DIR/README.md"
  fi
  log_done "terrascan installed"
}

# Function to install kubeaudit
install_kubeaudit() {
  log_info "Installing kubeaudit"
  if ! command -v kubeaudit &> /dev/null; then
    run_cmd "curl -Lo $WORKSHOP_DIR/kubeaudit.tar.gz https://github.com/Shopify/kubeaudit/releases/download/v0.22.2/kubeaudit_0.22.2_linux_amd64.tar.gz"
    run_cmd "tar -zxvf $WORKSHOP_DIR/kubeaudit.tar.gz -C $WORKSHOP_DIR"
    run_cmd "mv $WORKSHOP_DIR/kubeaudit /usr/local/bin/"
    run_cmd "chmod +x /usr/local/bin/kubeaudit"
    run_cmd "rm $WORKSHOP_DIR/kubeaudit.tar.gz $WORKSHOP_DIR/README.md"
  fi
  log_done "kubeaudit installed"
}

# Function to install nuclei
install_nuclei() {
  log_info "Installing nuclei"
  if ! command -v nuclei &> /dev/null; then
    run_cmd "curl -Lo $WORKSHOP_DIR/nuclei_3.3.9_linux_amd64.zip https://github.com/projectdiscovery/nuclei/releases/download/v3.3.9/nuclei_3.3.9_linux_amd64.zip"
    run_cmd "unzip $WORKSHOP_DIR/nuclei_3.3.9_linux_amd64.zip -d $WORKSHOP_DIR"
    run_cmd "mv $WORKSHOP_DIR/nuclei /usr/local/bin/"
    run_cmd "chmod +x /usr/local/bin/nuclei"
    run_cmd "rm $WORKSHOP_DIR/nuclei_3.3.9_linux_amd64.zip $WORKSHOP_DIR/README_CN.md $WORKSHOP_DIR/README_ID.md $WORKSHOP_DIR/README_KR.md $WORKSHOP_DIR/LICENSE.md $WORKSHOP_DIR/README_ES.md $WORKSHOP_DIR/README_JP.md $WORKSHOP_DIR/README.md"
  fi
  log_done "nuclei installed"
}

install_all_tools() {
  install_docker
  install_kind
  install_kubectl
  install_kustomize
  install_k9s
  install_mkcert
  install_kube_hunter
  install_kube_linter
  install_terrascan
  install_kubeaudit
  install_nuclei
}

install_selected_tools() {
  for tool in "${INSTALL_TOOLS[@]}"; do
    tool=$(echo "$tool" | xargs) # Trim whitespace
    case $tool in
      docker) install_docker ;;
      kind) install_kind ;;
      kubectl) install_kubectl ;;
      kustomize) install_kustomize ;;
      k9s) install_k9s ;;
      mkcert) install_mkcert ;;
      kube_hunter) install_kube_hunter ;;
      kube_linter) install_kube_linter ;;
      terrascan) install_terrascan ;;
      kubeaudit) install_kubeaudit ;;
      nuclei) install_nuclei ;;
      *)
        printf "Unknown tool: %s\n" "$tool"
        show_help
        exit 1
        ;;
    esac
  done
}

# Function to perform cleanup
cleanup() {
  log_info "Running apt autoremove"
  run_cmd "apt autoremove -y"
  log_done "apt autoremove completed"
}

# Main function to install all tools
main() {
  # Show help if no arguments are provided
  if [[ $INSTALL -eq 0 ]]; then
    show_help
    exit 0
  fi

  update_package_lists
  install_prerequisites

  if [[ ${#INSTALL_TOOLS[@]} -eq 0 ]]; then
    install_all_tools
  else
    install_selected_tools
  fi

  cleanup
}

main "$@"
