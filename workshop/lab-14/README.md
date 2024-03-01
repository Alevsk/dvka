# kube-hunter: Hunt for security weaknesses in Kubernetes clusters

## Quick Start

1. Installation

    Install on your system

    ```bash
    pip install kube-hunter
    kube-hunter
    ```

    **OR**

    Run via `docker container`

    ```bash
    docker run -it --rm --network host aquasec/kube-hunter
    ```

2. Run `kube-hunter` scanner outside the cluster

    ```bash
    # list scanning capabilities
    kube-hunter --list
    # scan local k8s cluster running via kind
    kube-hunter --kubeconfig="~/.kube/config" --k8s-auto-discover-nodes
    ```

3. Run `kube-hunter` scanner outside the cluster using a `service account`

    ```bash
    # create service account
    kubectl apply -f kube-hunter-sa.yaml
    # export service account to environment variable
    export KHTOKEN=$(kubectl get secrets kube-hunter-secret -o json | jq ".data.token" -j | base64 -d)
    # run kube-hunter
    kube-hunter --kubeconfig="~/.kube/config" --k8s-auto-discover-nodes --service-account-token=$KHTOKEN
    ```

4. Run `kube-hunter` scanner inside the cluster as pod

    ```bash
    # deploy kube-hunter pod
    kubectl apply -f kube-hunter.yaml
    # analyze pod logs after scan is completed
    kubectl logs -l app=kube-hunter --tail=-1
    ```

5. Run `kube-hunter` scanner inside the cluster as pod using a `service account`

    ```bash
    # deploy kube-hunter pod
    kubectl apply -f kube-hunter-with-sa.yaml
    # analyze pod logs after scan is completed
    kubectl logs -l app=kube-hunter-with-sa --tail=-1
    ```

6. Finalize the lab

    ```bash
    # end the lab
    kubectl delete -f kube-hunter.yaml
    kubectl delete -f kube-hunter-with-sa.yaml
    kubectl delete -f kube-hunter-sa.yaml
    ```

## Resouces

- <https://github.com/aquasecurity/kube-hunter>
