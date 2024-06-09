# cert-manager: X.509 certificate management for Kubernetes

## Quick Start

1. Install `cert-manager` using the default yaml configuration

   ```bash
   kubectl apply -f cert-manager.yaml
   ```

2. Provide or generate your own `root` Certificate Authority (CA), ie:

    ```bash
    # generate private key
    openssl genrsa -out rootCAKey.pem 2048
    # generate public key
    openssl req -x509 -sha256 -new -nodes -key rootCAKey.pem -days 3650 -out rootCACert.pem
    ```

3. `base64` encode the public and private keys for your root CA

   ```bash
   cat rootCACert.pem | base64 -w 0 # copy to clipboard
   cat rootCAKey.pem | base64 -w 0 # copy to clipboard
   ```

4. Open the `custom-ca-secret.yaml` file and place the `base64 encoded` values for the public and private key in the `tls.crt` and `tls.key` fields, then create create the custom ca on kubernetes

    ```bash
    # create secret
    kubectl apply -f custom-ca-secret.yaml
    # create cert-manager ClusterIssuer
    kubectl apply -f custom-ca.yaml
    ```

5. Verify the ca was added to the cluster issuers list

   ```bash
   kubectl get clusterissuers
   ```

6. Generate a new `tls` certificate using `cert-manager` and your root CA via `cert-manager`

   ```bash
   kubectl apply -f default-tls-certificate.yaml
   ```

7. Verify the certificate was generated correctly

   ```bash
   # check status of the certificate request
   kubectl get certificaterequests
   # check status of the certificate itself
   kubectl get certificates
   # check the tls certificate stored directly on the k8s secret
   kubectl describe secrets default-tls-certificate-secret describe
   # inspect the public and private keys of the generated certificate
   kubectl get secrets default-tls-certificate-secret -o yaml
   ```

   You can use <https://www.sslshopper.com/certificate-decoder.html> or <https://www.sslchecker.com/certdecoder> to verify the content of the certificate

8. Use this keypair to configure TLS for your workloads, similar to what you did for the [Configmaps & Secrets](../configmaps-secrets/README.md) lab

9. End the lab

   ```bash
   kubectl delete secret default-tls-certificate-secret
   kubectl delete -f default-tls-certificate.yaml
   kubectl delete -f custom-ca-secret.yaml
   kubectl delete -f custom-ca.yaml
   kubectl delete -f cert-manager.yaml
   ```

## Resouces

- <https://cert-manager.io/docs/installation/>
- <https://cert-manager.io/docs/tutorials/acme/nginx-ingress/>
