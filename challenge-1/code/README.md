# Development Guide: Hack The NFT Museum

This guide provides technical instructions for setting up and deploying the NFT Museum challenge.

## Requirements

- A running Kubernetes cluster ([Minikube](https://minikube.sigs.k8s.io/docs/start/) or [Kind](https://kind.sigs.k8s.io/))
- [Kustomize](https://kustomize.io/)
- Go 1.17 or newer
- Make

## Deployment

### Kubernetes

1.  **Configure the Challenge**

    -   Edit `k8s/base/secret.yaml` and replace `<REPLACE WITH SECRET HERE>` with a random string.
    -   Modify `k8s/base/deployment.yaml` and replace `<REPLACE WITH FLAG HERE>` with the flag value.

2.  **Deploy to the Cluster**

    ```bash
    # Apply the configuration
    kustomize build k8s/base | kubectl apply -f -

    # Expose the service
    kubectl port-forward svc/nft-store 8080:8080 -n lab-1
    ```

### Local Development

```bash
# Compile and run the application
make && ./cmd/app/lab1
```

### Docker

```bash
# Build and run the Docker container
make docker && docker run --rm -p 8080:8080 --name=dvka-labl-1 alevsk/dvka:lab-1

# Alternatively, use Docker Compose
docker-compose up -d
```

## Accessing the Application

Once deployed, access the NFT Museum at <http://localhost:8080/>.

## Environment Variables

-   `DVKA_LAB1_SIGNING_KEY`: JWT signing key for authentication.
-   `DVKA_LAB1_FLAG`: Challenge flag value.

## Cleanup

### Kubernetes

```bash
# Stop port forwarding and remove the deployment
ctrl-c

kustomize build k8s/base | kubectl delete -f -
```

### Local Development

Stop the application with `ctrl-c`.

### Docker

-   For `docker run`, stop the container with `ctrl-c`.
-   For Docker Compose, run `docker-compose down`.