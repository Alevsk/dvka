# LoadBalancer services with METALLB

`LoadBalancer` type services exposes services externally using an external load balancer. Kubernetes does not directly offer a load balancing component, for this example we are going to use `METALLB`.

## Quick Start

1. Install `metallb` on the cluster

    ```bash
    kubectl apply -f metallb-native.yaml -n metallb-system
    ```

2. Configure the `IPAddressPool` range. These are the `IP` addresses that will be assigned to each new `LoadBalancer` type service

   > **NOTE:** Make sure `metallb-configmap.yaml` contains the correct `IP` address range assigned to your `k8s` nodes, if you are running `kubernetes` via kind, these will be your container `IP` addresses, ie: `docker ps -aq | xargs docker inspect | jq '.[].NetworkSettings.Networks.kind.IPAddress' | sort`

   ```bash
   # apply configuration
   kubectl apply -f metallb-configmap.yaml -n metallb-system
   ```

3. Deploy `ingress-nginx` in the cluster via `kustomize`

   ```bash
   kustomize build ingress-nginx | kubectl apply -f -
   ```

4. Confirm a new `LoadBalancer` type service has been created an a `EXTERNAL-IP` address from the `IPAddressPool` has been assigned.

    ```bash
    kubectl get svc -n ingress-nginx
        
    NAME                                 TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)                                      AGE
    ingress-nginx-controller             LoadBalancer   10.98.139.45    172.19.0.2    80:30446/TCP,443:32721/TCP,10254:31726/TCP   17m
    ingress-nginx-controller-admission   ClusterIP      10.102.152.26   <none>        443/TCP                                      17m
    ```

5. Deploy example services and ingress rules

    ```bash
    kubectl apply -f foo-bar-services.yaml
    ```

6. Test `HTTP` requests against the externally exposed `foo` and `bar` services

    ```bash
    # query foo service
    curl http://172.19.0.2/foo
    <h1>foo</h1>
    # query bar service
    curl http://172.19.0.2/bar
    <h1>bar</h1>
    ```

7. Clean the environment

    ```bash
    kubectl delete -f foo-bar-services.yaml
    kustomize build ingress-nginx | kubectl delete -f -
    kubectl delete -f metallb-configmap.yaml -n metallb-system
    kubectl delete -f metallb-native.yaml -n metallb-system
    ```

## Resouces

- <https://metallb.io/>
- <https://medium.com/groupon-eng/loadbalancer-services-using-kubernetes-in-docker-kind-694b4207575d>
- <https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer>
