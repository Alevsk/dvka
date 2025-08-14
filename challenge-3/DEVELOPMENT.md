# Development Guide - Red Teaming Kubernetes Challenge

This guide provides technical instructions for setting up and deploying the Red Teaming Kubernetes challenge in various environments.

## Requirements

* A running Kubernetes cluster ([Minikube](https://minikube.sigs.k8s.io/docs/start/) or [Kind](https://kind.sigs.k8s.io/))
* [Kustomize](https://kustomize.io/)
* [kubectl](https://kubernetes.io/docs/tasks/tools/)

## Getting Started

1. Clone the repository:

```bash
git clone https://github.com/Alevsk/dvka && cd dvka/challenge-3
```

## Deployment Options

### 1. Kubernetes Deployment (Production-like)

1. Configure challenge parameters:

   * Edit `k8s/base/dvwa/secret.yaml` and replace `<CHALLENGE-3-FLAG-1-HERE>` with a flag value
   * Edit `k8s/base/ingress-nginx/controller-deployment.yaml` and replace `<CHALLENGE-3-FLAG-2-HERE>` with a flag value
   * Edit `k8s/base/echo-server/secret.yaml` and replace `<CHALLENGE-3-FLAG-3-HERE>` with a flag value
   * Edit `k8s/base/echo-server/deployment.yaml` and replace `<CHALLENGE-3-FLAG-4-HERE>` with a flag value

2. Deploy to cluster:

```bash
# Apply the configuration
kustomize build k8s/base | kubectl apply -f -

# Expose the vulnerable DVWA service to localhost:8080
kubectl port-forward svc/dvwa -n dvwa 8080:80

# Expose the ArgoCD UI to localhost:8081
kubectl port-forward svc/argocd-server -n argocd 8081:80
```
