# Network Security Policies With Calico

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

4. Exec into the running container

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

5. Install `curl` on both nginx containers

    - `apk add curl`

6. Test connectivity between services in two different namespaces

    From `tenant-1` to `tenant-2`

    ```bash
    curl http://nginx.tenant-2.svc.cluster.local:8080
    ```

    From `tenant-2` to `tenant-1`

    ```bash
    curl http://nginx.tenant-1.svc.cluster.local:8080
    ```

    Notice how the service `URLs` have the following structure: `http://<service name>.<namespace>.svc.cluster.local:<port>`

7. Install the Tigera Calico operator and custom resource definitions.

    ```bash
    kubectl create -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.1/manifests/tigera-operator.yaml
    ```

8. Install Calico by creating the necessary custom resource.

    > Before creating this manifest, read its contents and make sure its settings are correct for your environment. For example, you may need to change the default IP pool CIDR to match your pod network CIDR. See <https://docs.tigera.io/calico/latest/getting-started/kubernetes/quickstart>.

    ```bash
    kubectl create -f custom-resources.yaml
    ```

9. Confirm that all of the pods are running with the following command.

    ```bash
    watch kubectl get pods -n calico-system
    ```

10. Block all incomming request to `tenant-2` namespace workloads using Calico Networking Policies. Look at `np-default-deny.yaml` and then run:

    > **Note:** In some cluster setups nginx pods need to be restarted first.

    ```bash
    kubectl apply -f np-default-deny.yaml
    ```

11. Exec into the `tenant-1` running container again and test connectivity to the running container in `tenant-2`. You should see a timeout error.

    ```bash
    curl -m 5 http://nginx.tenant-2.svc.cluster.local:8080
    ```

12. Exec into the `tenant-2` running container again and test connectivity to the running container in `tenant-1`.

    ```bash
    # request to tenant-1
    curl -m 5 http://nginx.tenant-1.svc.cluster.local:8080
    # request to itself using service name should timeout
    curl -m 5 http://nginx.tenant-2.svc.cluster.local:8080
    # request to itself using localhost
    curl -m 5 http://localhost:8080
    ```

13. Deploy networking rule to allow internal connectivity for `tenant-2` namespace.

    ```bash
    kubectl apply -f np-allow-namespace-connectivity.yaml  
    ```

    Exec into the `tenant-2` running container again and test connectivity

    ```bash
    # request to itself using service name should work this time
    curl -m 5 http://nginx.tenant-2.svc.cluster.local:8080
    ```

14. Deploy a new tenant namespace. Update existing networking rule to allow connectivity from workloads running on `tenant-3` namespace to `tenant-2` namespace.

    ```bash
    # deploy tenant-3
    kubectl apply -f tenant-3.yaml
    # update tenant-2 networking rule to accept tenant-3 connections
    kubectl apply -f np-allow-namespace-connectivity-update.yaml  
    ```

    Exec into the `tenant-3` running container

    **kubectl:**

    ```bash
    # exec into nginx tenant-3
    kubectl -n tenant-3 exec -it <pod name> -- sh
    ```

    **k9s:**

    - Namespace > `tenant-3` > Pods `>` nginx `>` press `<s>`

    Install `curl` on `tenant-3` container and test connectivity to `tenant-2`.

    ```bash
    # install curl
    apk add curl
    # request to tenant-2 should work this time
    curl -m 5 http://nginx.tenant-2.svc.cluster.local:8080
    ```

15. Finalize the lab

    ```bash
    # end the lab
    kubectl delete -f tenant-1.yaml
    kubectl delete -f tenant-2.yaml
    kubectl delete -f tenant-3.yaml
    kubectl delete -f np-default-deny.yaml
    kubectl delete -f np-allow-namespace-connectivity.yaml
    kubectl delete -f np-allow-namespace-connectivity-update.yaml
    kubectl delete -f custom-resources.yaml
    kubectl delete -f https://raw.githubusercontent.com/projectcalico/calico/v3.26.1/manifests/tigera-operator.yaml
    ```

## Resources

- <https://docs.tigera.io/calico/latest/network-policy/get-started/kubernetes-policy/kubernetes-policy-basic#enable-isolation>
- <https://docs.tigera.io/calico/latest/reference/resources/networkpolicy>
- <https://docs.tigera.io/calico/latest/network-policy/policy-rules/namespace-policy>
