# Pod Resource Limits

## Quick Start

1. Install metrics server in your cluster

    ```bash
    kubectl apply -f metrics-server.yaml
    ```

    After the service is running restart `k9s` and look for the new `CPU` and `Memory` metrics available for your cluster. Look at metrics using `kubectl`:

    ```bash
    kubectl top pod -A
    ```

2. Memory testing

    ```bash
    # deploy container without limits
    kubectl apply -f mem-testing.yaml
    # look at memory consumtion by the pod
    watch kubectl top pod -l app=mem-testing
    ```

    Update memory stress container to limit the amount of memory to only 100mb

    ```bash
    # deploy container with limits
    kubectl apply -f mem-testing-limits.yaml
    ```

    Observe how the container is stopped with status `OOMKilled`

3. CPU testing

    ```bash
    # deploy container without limits
    kubectl apply -f cpu-testing.yaml
    # look at cpu consumtion by the pod
    watch kubectl top pod -l app=cpu-testing
    ```

    > The `-cpus "2"` argument tells the Container to attempt to use 2 CPUs.

    Update cpu stress container to limit the amount of cpu to only 1 core

    ```bash
    # deploy container with limits
    kubectl apply -f cpu-testing-limits.yaml
    ```

    Observe how the container is limited to consume maximum 1 cpu

4. Finalize the lab

    ```bash
    kubectl delete -f metrics-server.yaml
    kubectl delete -f mem-testing.yaml
    kubectl delete -f cpu-testing.yaml
    ```

## Resouces

- <https://kubernetes.io/docs/tasks/configure-pod-container/assign-memory-resource/#exceed-a-container-s-memory-limit>
