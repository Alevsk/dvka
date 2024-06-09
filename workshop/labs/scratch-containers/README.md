# Scratch Containers

## Quick Start

1. Take a look at the example application source code under `./encoding-service`

    * main.go
    * base64
    * Dockerfile

2. Build the `encoding service` docker image

    ```bash
    docker build -t alevsk/dvka:lab10 -f encoding-service/Dockerfile ./encoding-service
    ```

3. Push the image to your Kubernetes cluster

    ```bash
    kind load docker-image alevsk/dvka:lab10 --name workshop-cluster
    ```

4. Deploy the application into Kubernetes

    ```bash
    # create deployment
    kubectl apply -f encoding-service.yaml
    # locally expose the application service
    kubectl port-forward svc/encoding-service 1337:1337
    ```

5. Open the browser and go to <http://localhost:1337/run?command=encode&message=hello%20world>

    * Found any vulnerabilities?
    * Get a shell on the container using `kubectl` or `k9s`

6. Stop `port-forward` (<ctrl+c>) and remove application

    ```bash
    kubectl delete -f encoding-service.yaml
    ```

7. Build the scratch container and deploy to kubernetes again

    ```bash
    # build scratch image
    docker build -t alevsk/dvka:lab10-scratch -f encoding-service/Dockerfile.scratch ./encoding-service
    # push image to kubernetes
    kind load docker-image alevsk/dvka:lab10-scratch --name workshop-cluster
    # create deployment
    kubectl apply -f encoding-service-scratch.yaml
    # locally expose the application service
    kubectl port-forward svc/encoding-service 1337:1337
    ```

8. Open the browser and go to <http://localhost:1337/run?command=encode&message=hello%20world>

    * Test for vulnerabilities again
    * Get a shell on the container using `kubectl` or `k9s`

9. Stop `port-forward` (<ctrl+c>) and remove application

    ```bash
    kubectl delete -f encoding-service.yaml
    ```

10. Follow up

    * Differences between regular images and scratch images

## Resouces

* <https://github.com/GoogleContainerTools/distroless>
* <https://hub.docker.com/_/scratch>
* <https://hub.docker.com/r/redhat/ubi8/tags>
