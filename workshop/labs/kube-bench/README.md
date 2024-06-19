# Kube-bench: CIS Kubernetes Benchmark

## Quick Start

Checks whether Kubernetes is deployed according to security best practices as defined in the [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)

1. Deploy `kube-bench` into your cluster do start the assessment.

    ```bash
    kubectl apply -f kube-bench.yaml
    ```

2. Confirm `kube-bench` pod was created and status is `Completed`.

    ```bash
    kubectl get pods
    ```

3. Inspect `kube-bench` report in pod logs

    **kubectl:**

    ```bash
    kubectl logs -l app=kube-bench --tail=-1
    ```

    **k9s:**

    - Pods `>` kube-bench `>` press `<l>`

4. Finalize the lab

    ```bash
    # end the lab
    kubectl delete -f kube-bench.yaml
    ```

## Resources

- <https://github.com/aquasecurity/kube-bench>
- <https://www.cisecurity.org/benchmark/kubernetes>
