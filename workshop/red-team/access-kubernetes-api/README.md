# Access Kubernetes API server

The Kubernetes API server is the gateway to the cluster. Actions in the cluster are performed by sending various requests to the RESTful API. The status of the cluster, which includes all the components that are deployed on it, can be retrieved by the API server. Attackers may send API requests to probe the cluster and get information about containers, secrets, and other resources in the cluster.

In addition, the Kubernetes API server can also be used to query information about Role Based Access (RBAC) information such as Roles, ClusterRoles, RoleBinding, ClusterRoleBinding and Service Accounts. Attacker may use this information to discover permissions and access associated with Service Accounts in the cluster and use this information to progress towards its attack objectives.

## Resources
