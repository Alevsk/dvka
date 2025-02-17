# Namespaces

## Quick Start

1. Look at `tenant-1.yaml` file and deploy all the resources for application 1

    ```bash
    kubectl apply -f tenant-1.yaml
    ```

2. Look at `tenant-2.yaml` file and deploy all the resources for application 2

    ```bash
    kubectl apply -f tenant-2.yaml
    ```

3. Inspect the resources created for the `tenant-1` and `tenant-2` namespaces using `k9s` or `kubectl`

    ```bash
    # tenant-1
    kubectl get all --namespace tenant-1
    # tenant-2
    kubectl get all --namespace tenant-2
    ```

4. Start `port-forward` for both applications

    ```bash
    # forwarding tenant-1 in first terminal
    kubectl port-forward svc/nginx 8081:8080 -n tenant-1
    # forwarding tenant-2 in second terminal
    kubectl port-forward svc/nginx 8082:8080 -n tenant-2
    ```

    Open browser and go to <http://localhost:8081> and <http://localhost:8082> to verify applications are running correctly

5. Exec into the running container

    **kubectl:**

    ```bash
    # exec into nginx tenant-1
    kubectl -n tenant-1 exec -it <pod name> -- sh
    # exec into nginx tenant-2
    kubectl -n tenant-2 exec -it <pod name> -- sh
    ```

    **k9s:**

    - Namespace > `tenant-1` > Pods `>` nginx `>` press `<s>`
    - Namespace > `tenant-2` > Pods `>` nginx `>` press `<s>`

6. Install `curl` on both nginx containers

    - `apk add curl`

7. Test connectivity between services in two different namespaces

    From `tenant-1` to `tenant-2`

    ```bash
    curl http://nginx.tenant-2.svc.cluster.local:8080
    ```

    From `tenant-2` to `tenant-1`

    ```bash
    curl http://nginx.tenant-1.svc.cluster.local:8080
    ```

    Notice how the service `URLs` have the following structure: `http://<service name>.<namespace>.svc.cluster.local:<port>`

8. Stop `port-forward` (<ctrl+c>) and remove applications:

    ```bash
    kubectl delete -f tenant-1.yaml
    kubectl delete -f tenant-2.yaml
    ```

## Resources

- <https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/>
