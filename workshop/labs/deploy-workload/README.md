# Deploy Kubernetes Workload

## Using Multiple Kubectl Commands

1. Run the following commands

    ```bash
    # create nginx deployment
    kubectl create deployment nginx --image=nginx:stable-alpine3.17-slim --replicas=2 --port=80
    # create service (nginx by default will run in port 80)
    kubectl create service clusterip nginx --tcp=8080:80
    # locally expose nginx service using port 8080
    kubectl port-forward svc/nginx 8080:8080
    ```

2. Open browser and go to <http://localhost:8080/>

3. Explore deployed application using `kubectl` or `k9s`

4. Terminate nginx application

    ```bash
    kubectl delete svc nginx
    kubectl delete deployment nginx
    ```

## Using Yaml Files

1. Run the following commands

    ```bash
    # create nginx deployment and service (nginx by default will run in port 80)
    kubectl apply -f nginx.yaml
    # locally expose nginx service using port 8080
    kubectl port-forward svc/nginx 8080:8080
    ```

2. Open browser and go to <http://localhost:8080/>

3. Explore deployed application using `kubectl` or `k9s`

    - Pods
    - Deployments
    - Services

4. Terminate nginx application

    ```bash
    kubectl delete -f nginx.yaml
    ```

## Resources

- <https://kubernetes.io/docs/tasks/run-application/run-stateless-application-deployment/>
