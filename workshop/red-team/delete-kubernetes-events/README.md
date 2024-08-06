# Delete Kubernetes events

A Kubernetes event is a Kubernetes object that logs state changes and failures of the resources in the cluster. Example events are a container creation, an image pull, or a pod scheduling on a node.

Kubernetes events can be very useful for identifying changes that occur in the cluster. Therefore, attackers may want to delete these events (e.g., by using: `kubectl delete events–all`) in an attempt to avoid detection of their activity in the cluster.

## Resources
