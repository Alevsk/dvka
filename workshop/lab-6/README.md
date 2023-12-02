# Configmaps & Secrets

## Configmap

1. Create a new `configmap` from literal

    ```bash
    kubectl create configmap lab-6-configmap --from-literal=workshop="kubernetes security" --from-literal=lab="Lab 6"
    ```

2. Look at lab-6-configmap `configmap` using `k9s` or `kubectl`

3. Create new `configmaps` from a file

    ```bash
    kubectl create configmap nginx-configuration --from-file=default.conf=default.conf
    kubectl create configmap nginx-index --from-file=index.html=index.html
    ```

4. Look at nginx-configuration `configmap` using `k9s` or `kubectl`

5. Deploy nginx using a custom `default.conf` configuration

    ```bash
    # create nginx deployment and service (nginx will run in port 8080)
    kubectl apply -f nginx.yaml
    # locally expose nginx service
    kubectl port-forward svc/nginx 8080:8080
    ```

6. Open browser and go to <http://localhost:8080/>

7. Navigate inside the nginx pod and look for the `default.conf` and `index.html` files

## Secret

1. Create a new `secret` from literal

    ```bash
    kubectl create secret generic lab-6-secret --from-literal=password=1234567
    ```

2. Look at lab-6-secret `secret` using `k9s` or `kubectl`

3. Create new `secret` from a file

    > Generate self-signed tls certificates using [mkcert](https://github.com/FiloSottile/mkcert), openssl or any other tool you want.<br>
    > ie:<br>
    > `mkcert localhost`

    ```bash
    kubectl create secret tls nginx-tls-certificates --cert=localhost.pem --key=localhost-key.pem
    ```

4. Create new `configmaps` from `nginx` configuration that include tls certificates

    ```bash
    kubectl create configmap nginx-configuration-tls --from-file=default.conf=default-tls.conf
    ```

5. Deploy nginx with tls certificates

    ```bash
    # create nginx deployment and service (nginx will run in port 8443)
    kubectl apply -f nginx-tls.yaml
    # locally expose nginx service
    kubectl port-forward svc/nginx 8443:8443
    ```

6. Open browser and go to <https://localhost:8443/> or use `curl https://localhost:8443 -v`

7. Navigate inside the nginx pod and look for the `default.conf` and the tls certificate files

## End the Lab

Stop `port-forward` (<ctrl+c>) and remove application files

```bash
kubectl delete -f nginx.yaml
kubectl delete -f nginx-tls.yaml
kubectl delete configmap lab-6-configmap 
kubectl delete configmap nginx-configuration
kubectl delete configmap nginx-configuration-tls
kubectl delete configmap nginx-index
kubectl delete secret lab-6-secret
kubectl delete secret nginx-tls-certificates
```

## Resouces

- <https://kubernetes.io/docs/concepts/configuration/configmap/>
- <https://kubernetes.io/docs/tasks/configure-pod-container/configure-pod-configmap/>
- <https://kubernetes.io/docs/concepts/configuration/secret/>
