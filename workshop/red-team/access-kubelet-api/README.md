# Access Kubelet API

Kubelet is the Kubernetes agent that is installed on each node. Kubelet is responsible for the proper execution of pods that are assigned to the node. Kubelet exposes a read-only API service that does not require authentication (TCP port 10255). Attackers with network access to the host (for example, via running code on a compromised container) can send API requests to the Kubelet API. Specifically querying `https://[NODE IP]:10255/pods/` retrieves the running pods on the node. `https://[NODE IP]:10255/spec/` retrieves information about the node itself, such as CPU and memory consumption.

## Resources

- <https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/>
