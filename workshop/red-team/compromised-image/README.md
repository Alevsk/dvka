# Compromised Image in Registry

An attacker who can push to an image registry — or who can trick an operator into pulling a malicious public image — can run arbitrary code inside the cluster the moment the image is scheduled. The backdoor is baked into an otherwise legitimate-looking container layer.

## Description

Running a compromised image in a cluster can compromise the cluster. Attackers who get access to a private registry can plant their own compromised images in the registry. Those images are then pulled by unsuspecting users or automated CD pipelines. In addition, developers frequently use untrusted images from public registries (such as Docker Hub) that may already be malicious or may be subject to a typosquatting attack.

The attack works in two phases:

1. **Build phase**: The attacker starts from a trusted base image and adds a hidden layer — a startup script, a modified entrypoint, or a compiled binary — that harvests credentials, steals the Kubernetes service account token, or opens a reverse shell.
2. **Runtime phase**: The container starts, looks completely normal from the outside (the legitimate application still runs), but the malicious payload executes silently in parallel.

This scenario simulates the runtime phase directly inside a Kind cluster using a Deployment whose entrypoint mimics what a backdoored image would do.

## Prerequisites

- A running Kubernetes cluster (Kind `workshop-cluster` is assumed).
- `kubectl` installed and configured to connect to your cluster.
- `docker` installed locally (for the optional image-build walkthrough).

## Quick Start

### 1. Understand the backdoored Dockerfile

Review the example `Dockerfile` and `backdoor.sh` in this directory. The Dockerfile adds a script to nginx's `/docker-entrypoint.d/` directory. When the container starts, nginx's official entrypoint runs every script in that directory before launching the server.

```bash
cat Dockerfile
cat backdoor.sh
```

> **To build and push your own test image** (requires a registry you control):
>
> ```bash
> docker build -t YOUR_REGISTRY/nginx-backdoored:1.25 .
> docker push YOUR_REGISTRY/nginx-backdoored:1.25
> ```
>
> Then update the `image:` field in `backdoored-app.yaml` to point to your registry.

### 2. Deploy the scenario

The provided `backdoored-app.yaml` uses the standard `nginx:1.25-alpine` image but overrides the entrypoint to reproduce the exact behavior a backdoored image would exhibit — credential harvesting at startup, followed by launching the legitimate server.

```bash
kubectl apply -f backdoored-app.yaml
```

Wait for the pod to be running:

```bash
kubectl wait --for=condition=Ready pod -l app=legitimate-app -n compromised-image --timeout=60s
```

### 3. Observe the backdoor executing at startup

Check the container logs immediately after startup to see the backdoor output:

```bash
kubectl logs -l app=legitimate-app -n compromised-image
```

Expected output:

```
[BACKDOOR] Exfiltrating environment variables...
[BACKDOOR] Dumping service account token...
[BACKDOOR] Data staged at /tmp/exfil.txt
[BACKDOOR] Starting legitimate process...
```

The nginx server is running normally. An operator checking the service would see no anomaly.

### 4. Inspect the staged exfiltration data

Exec into the pod and read what the backdoor collected:

```bash
kubectl exec -it deploy/legitimate-app -n compromised-image -- cat /tmp/exfil.txt
```

The file contains all environment variables — including the injected `DB_PASSWORD` and `API_KEY` — plus the Kubernetes service account token. In a real attack, this data would already be on the attacker's server.

### 5. Use the stolen service account token

The `backdoored-app.yaml` grants the `default` service account in the `compromised-image` namespace read access to cluster resources via a ClusterRoleBinding. This simulates the over-privileged service account that is common in real environments.

Inside the pod, use the harvested token to query the Kubernetes API:

```bash
kubectl exec -it deploy/legitimate-app -n compromised-image -- /bin/sh
```

Inside the container:

```bash
TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)
CACERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt

# Query the API server using the stolen token — list all namespaces
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  https://kubernetes.default.svc.cluster.local/api/v1/namespaces | grep '"name"'

# List all pods across namespaces
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  https://kubernetes.default.svc.cluster.local/api/v1/pods | grep '"name"' | head -20
```

### 6. Inspect image layers to detect the backdoor (defender perspective)

To understand how defenders can catch this, inspect the image history locally:

```bash
# Pull the image and inspect its layers
docker pull nginx:1.25-alpine
docker history nginx:1.25-alpine

# With a real backdoored image, look for unexpected COPY or RUN layers
# that reference scripts or executables not in the original image.
docker inspect nginx:1.25-alpine | python3 -m json.tool | grep -A5 "Layers"
```

Tools like [Trivy](https://github.com/aquasecurity/trivy), [Grype](https://github.com/anchore/grype), and [Docker Scout](https://docs.docker.com/scout/) can detect known malicious layers and suspicious additions in CI pipelines.

## Cleanup

```bash
kubectl delete -f backdoored-app.yaml
```

## Resources

- [Supply Chain Threats Using Container Images](https://blog.aquasec.com/supply-chain-threats-using-container-images)
- [Malicious Docker Hub Container Images Cryptojacking](https://www.trendmicro.com/vinfo/us/security/news/virtualization-and-cloud/malicious-docker-hub-container-images-cryptocurrency-mining)
- [MITRE ATT&CK: Supply Chain Compromise](https://attack.mitre.org/techniques/T1195/)
- [Trivy: Container Image Scanner](https://github.com/aquasecurity/trivy)
- [CNCF Software Supply Chain Security Best Practices](https://project.linuxfoundation.org/hubfs/CNCF_SSCP_v1.pdf)
