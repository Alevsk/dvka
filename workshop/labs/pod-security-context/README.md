# Pod Security Context

## Quick Start

1. Deploy ubuntu pod

    ```bash
    # create ubuntu pod
    kubectl apply -f ubuntu.yaml
    ```

2. Exec into the running container

    **kubectl:**

    ```bash
    kubectl exec -it ubuntu -- sh
    ```

    **k9s:**

    Pods `>` ubuntu `>` press `<s>`

3. Run the `whoami` command and try to install some applications

    - `apt-get update && apt-get install curl -y`

4. Terminate pod

    ```bash
    kubectl delete -f ubuntu.yaml
    ```

5. Deploy pod with `security context` and exec into it

    ```bash
    kubectl apply -f ubuntu-with-security-context.yaml
    ```

6. Run the `whoami` command and try to update system packages

7. Terminate pod

    ```bash
    kubectl delete -f ubuntu-with-security-context.yaml
    ```

## Resources

- <https://kubernetes.io/docs/tasks/configure-pod-container/security-context/>
- <https://kubernetes.io/docs/concepts/security/pod-security-standards/>
