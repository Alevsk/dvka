# Pod or container name similarity

Pods that are created by controllers such as Deployment or DaemonSet have random suffix in their names. Attackers can use this fact and name their backdoor pods as they were created by the existing controllers. For example, an attacker could create a malicious pod named `coredns-{random suffix}` which would look related to the CoreDNS Deployment.

Also, attackers can deploy their containers in the kube-system namespace where the administrative containers reside.

## Quick Start

1. List `coredns` pods running in the cluster

    ```bash
    kubectl get pods -n kube-system -l k8s-app=kube-dns
    ```

    Observe how the pods have the following format `coredns-{random suffix}` in their names.

    Edit `pod.yaml` file. Set a name using a similar suffix, ie: `coredns-7db6d8ff4d-8adtw`

2. Create the new busybox pod by running the following command

    ```bash
    kubectl apply -f pod.yaml
    ```

3. List all the `coredns` pods again

    ```bash
    kubectl get pods -n kube-system -l k8s-app=kube-dns
    ```

    Your new `coredns` pod should be there, this pod in reality is running `busybox`

4. Terminate the pod

    ```bash
    kubectl delete -f pod.yaml
    ```

## Resources
