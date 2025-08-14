# Development Guide - Hack The NFT Museum Challenge

This guide provides technical instructions for setting up and deploying the NFT Museum challenge in various environments.

## Requirements

* A running Kubernetes cluster ([Minikube](https://minikube.sigs.k8s.io/docs/start/) or [Kind](https://kind.sigs.k8s.io/))
* [Kustomize](https://kustomize.io/)
* Go 1.17 or newer <https://go.dev/dl/>
* Make <https://www.gnu.org/software/make/>

## Getting Started

1. Clone the repository:

```bash
git clone https://github.com/Alevsk/dvka && cd dvka/challenge-1
```

## Deployment Options

### 1. Kubernetes Deployment (Production-like)

1. Configure challenge parameters:

   * Edit `k8s/base/secret.yaml` and replace `<REPLACE WITH SECRET HERE>` with a random string
   * Modify `k8s/base/deployment.yaml` and replace `<REPLACE WITH FLAG HERE>` with the flag value

2. Deploy to cluster:

```bash
# Apply the configuration
kustomize build k8s/base | kubectl apply -f -

# Expose the service
kubectl port-forward svc/nft-store 8080:8080 -n lab-1
```

### 2. Local Development (Development-like)

```bash
# Compile the binary
make && cd cmd/app

# Run the binary
DVKA_LAB1_SIGNING_KEY="" DVKA_LAB1_FLAG="" ./lab1
```

### 3. Docker Container (Development-like)

```bash
# Build docker image
TAG=alevsk/dvka:lab-1 make docker

# Run with Docker
docker run --rm -p 8080:8080 \
  -e DVKA_LAB1_SIGNING_KEY="" \
  -e DVKA_LAB1_FLAG="" \
  --name=dvka-labl-1 alevsk/dvka:lab-1

# Alternative: Using Docker Compose
docker-compose up -d
```

## Accessing the Application

Once deployed, access the NFT Museum at: <http://localhost:8080/>

## Environment Variables

* `DVKA_LAB1_SIGNING_KEY`: JWT signing key for authentication
* `DVKA_LAB1_FLAG`: Challenge flag value

## Cleanup Instructions

### Kubernetes

```bash
# Stop port forwarding
ctrl-c

# Remove deployment
kustomize build k8s/base | kubectl delete -f -
```

### Local Development

* Stop the binary with `ctrl-c`

### Docker

* For `docker run`: Stop with `ctrl-c`
* For Docker Compose: Run `docker-compose down`
