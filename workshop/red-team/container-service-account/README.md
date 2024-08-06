# Container service account

Service account (SA) represents an application identity in Kubernetes. By default, a Service Account access token is mounted to every created pod in the cluster and containers in the pod can send requests to the Kubernetes API server using the Service Account credentials. Attackers who get access to a pod can access the Service Account token (located in `/var/run/secrets/kubernetes.io/serviceaccount/token`) and perform actions in the cluster, according to the Service Account permissions. If RBAC is not enabled, the Service Account has unlimited permissions in the cluster. If RBAC is enabled, its permissions are determined by the RoleBindings \ ClusterRoleBindings that are associated with it.

An attacker which get access to the Service Account token can also authenticate and access the Kubernetes API server from outside the cluster and maintain access to the cluster.

## Quick Start

1. Deploy ubuntu pod

    ```bash
    # create ubuntu pod
    kubectl apply -f ubuntu.yaml
    ```

2. Exec into the running container

    **kubectl:**

    ```bash
    kubectl exec -it pod/ubuntu -- /bin/bash
    ```

    **k9s:**

    Pods `>` ubuntu `>` press `<s>`

3. Install `curl`

    - `apt-get update && apt-get install curl jq -y`

4. Move to the `serviceaccount` folder

    - `cd /var/run/secrets/kubernetes.io/serviceaccount`

5. Analyze the 3 files under the `serviceaccount` folder

    - ca.crt
    - namespace
    - token

6. Visualize `token` using <https://jwt.io/> or a similar tool

7. Query the the Kubernetes api server

    ```bash
    curl https://kubernetes.default.svc.cluster.local
    # ignore tls verification
    curl https://kubernetes.default.svc.cluster.local -k
    # pass ca.crt to verify tls connection
    curl https://kubernetes.default.svc.cluster.local --cacert ca.crt
    ```

8. Authenticate using the service account `token`

    ```bash
    export TOKEN=$(cat token)
    curl --cacert ca.crt https://kubernetes.default.svc.cluster.local/api/v1/namespaces?limit=500 -H "Authorization: Bearer $TOKEN"
    # Use jq to parse the list of existing namespaces in the cluster
    curl --cacert ca.crt https://kubernetes.default.svc.cluster.local/api/v1/namespaces?limit=500 -H "Authorization: Bearer $TOKEN" | jq ".items[].metadata.name"

    ```

9. Deploy ubuntu pod

    ```bash
    # delete ubuntu pod
    kubectl delete -f ubuntu.yaml
    # create ubuntu pod without mounting service account by default
    kubectl apply -f ubuntu-no-sa.yaml
    ```

10. Try to navigate again to the `serviceaccount` folder (You should get an error)

    - `kubectl exec -it pod/ubuntu -- /bin/bash`
    - `cd /var/run/secrets/kubernetes.io/serviceaccount`

11. Finalize the lab

    ```bash
    # end the lab
    kubectl delete -f ubuntu-no-sa.yaml
    kubectl delete -f ubuntu.yaml
    ```

## Resources

- <https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/>
- <https://kubernetes.io/docs/concepts/security/service-accounts/>
