#!/usr/bin/env bash
# setup-certs.sh
# Generates a self-signed TLS certificate for the webhook server,
# stores it as a Kubernetes Secret, and patches the caBundle in the
# MutatingWebhookConfiguration so the API server trusts the webhook.
#
# Usage:
#   chmod +x setup-certs.sh
#   ./setup-certs.sh

set -euo pipefail

NAMESPACE="webhook-system"
SERVICE="malicious-webhook"
SECRET_NAME="webhook-tls"
WEBHOOK_NAME="malicious-sidecar-injector"
CERT_DIR="$(mktemp -d)"

echo "[*] Generating TLS certificate for ${SERVICE}.${NAMESPACE}.svc ..."

# Generate CA key and self-signed certificate
openssl genrsa -out "${CERT_DIR}/ca.key" 2048 2>/dev/null
openssl req -new -x509 -days 365 \
  -key "${CERT_DIR}/ca.key" \
  -out "${CERT_DIR}/ca.crt" \
  -subj "/CN=webhook-ca/O=dvka" 2>/dev/null

# Generate server key and CSR
openssl genrsa -out "${CERT_DIR}/server.key" 2048 2>/dev/null
openssl req -new \
  -key "${CERT_DIR}/server.key" \
  -out "${CERT_DIR}/server.csr" \
  -subj "/CN=${SERVICE}.${NAMESPACE}.svc/O=dvka" 2>/dev/null

# Sign the server cert with the CA
cat > "${CERT_DIR}/san.cnf" <<EOF
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[v3_req]
subjectAltName = DNS:${SERVICE},DNS:${SERVICE}.${NAMESPACE},DNS:${SERVICE}.${NAMESPACE}.svc,DNS:${SERVICE}.${NAMESPACE}.svc.cluster.local
EOF

openssl x509 -req -days 365 \
  -in "${CERT_DIR}/server.csr" \
  -CA "${CERT_DIR}/ca.crt" \
  -CAkey "${CERT_DIR}/ca.key" \
  -CAcreateserial \
  -out "${CERT_DIR}/server.crt" \
  -extensions v3_req \
  -extfile "${CERT_DIR}/san.cnf" 2>/dev/null

echo "[*] Creating namespace ${NAMESPACE} (if not exists) ..."
kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

echo "[*] Storing TLS certificate as Secret ${SECRET_NAME} ..."
kubectl create secret tls "${SECRET_NAME}" \
  --namespace="${NAMESPACE}" \
  --cert="${CERT_DIR}/server.crt" \
  --key="${CERT_DIR}/server.key" \
  --dry-run=client -o yaml | kubectl apply -f -

echo "[*] Encoding CA bundle ..."
CA_BUNDLE=$(base64 < "${CERT_DIR}/ca.crt" | tr -d '\n')

echo "[*] Patching caBundle in MutatingWebhookConfiguration ..."
# Temporarily set the caBundle in the manifest
sed "s|caBundle: \"\"|caBundle: ${CA_BUNDLE}|g" \
  "$(dirname "$0")/mutating-webhook.yaml" | kubectl apply -f -

echo "[*] Cleaning up temp directory ..."
rm -rf "${CERT_DIR}"

echo ""
echo "[+] Setup complete. Deploy the webhook server next:"
echo "    kubectl apply -f webhook-server-code.yaml"
echo "    kubectl apply -f webhook-server.yaml"
