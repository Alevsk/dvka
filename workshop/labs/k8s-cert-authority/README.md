# Kubernetes certificate authority

## Quick Start

1. Generate a Private Key

    > "Elliptic Curve Digital Signature Algorithm" (ECDSA) with the P-256 curve. ECDSA is a widely-used and secure algorithm for generating key pairs.

    ```bash
    openssl ecparam -name prime256v1 -genkey -noout -out private.key
    ```

    Optionally you could use a different algorithm, ie: RSA.

    ```bash
    openssl genrsa -out private.key 2048
    ```

2. Generate a Certificate Signing Request (CSR)

   ```bash
   openssl req -new -config cert.cnf -key private.key -out kubernetes-security.csr
   ```

   > *Note:* Use <https://www.sslshopper.com/csr-decoder.html> to verify the generated `csr`

3. Encode the CSR in Base64 and copy it to your clipboard

   ```bash
   cat kubernetes-security.csr | base64 | tr -d "\n"
   ```

4. Open the `csr.yaml` file and paste the encoded certificate on the `spec.request` field if is not there already

5. Create the `CSR` resource in Kubernetes

   ```bash
   kubectl apply -f csr.yaml
   ```

6. Using `k9s` or `kubectl` list and inspect the created `CSR` resource

   ```bash
   kubectl get csr -A
   ```

7. Manually approve the CSR

   ```bash
   kubectl certificate approve kubernetes-security-csr
   ```

8. Retrieve the Signed Certificate

   ```bash
   kubectl get csr kubernetes-security-csr -o jsonpath='{.status.certificate}'| base64 -d > public.crt
   ```

   When the `CSR` is approved, the new certificate will be issued by the Kubernetes CA and would be found under the `status.certificate` field of the `kubernetes-security-csr` csr

9. Verify the `public.crt` certificate was issued by `kubernetes` using <https://www.sslchecker.com/certdecoder> or the `openssl` command

   ```bash
   cat public.crt | openssl x509 -noout -text
   ```

10. Use this keypair to configure TLS for your workloads, similar to what you did for the [Configmaps & Secrets](../configmaps-secrets/README.md) lab

## Resources

- <https://kubernetes.io/docs/reference/access-authn-authz/certificate-signing-requests/>
