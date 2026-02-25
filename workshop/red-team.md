# Red Team Techniques

This section provides an overview of common attack techniques used against Kubernetes clusters.

## Initial Access

- [Using Cloud Credentials](./red-team/using-cloud-credentials/README.md)
- [Compromised Image in Registry](./red-team/compromised-image/README.md)
- [Kubeconfig File](./red-team/kubeconfig-file/README.md)
- [Application Vulnerability](./red-team/application-vulnerability/README.md)
- [Exposed Sensitive Interfaces](./red-team/exposed-sensitive-interfaces/README.md)

## Execution

- [Exec into Container](./red-team/exec-into-container/README.md)
- [Bash/cmd inside Container](./red-team/bash-cmd-inside-container/README.md)
- [New Container](./red-team/new-container/README.md)
- [Application Exploit (RCE)](./red-team/application-exploit-rce/README.md)
- [SSH Server Running](./red-team/ssh-server-running/README.md)
- [Sidecar Injection](./red-team/sidecar-injection/README.md)

## Persistence

- [Backdoor Container](./red-team/backdoor-container/README.md)
- [Writable hostPath Mount](./red-team/writable-hostpath/README.md)
- [Kubernetes CronJob](./red-team/kubernetes-cronjob/README.md)
- [Malicious Admission Controller](./red-team/malicious-admission-controller/README.md)
- [Container Service Account](./red-team/container-service-account/README.md)
- [Static Pods](./red-team/static-pods/README.md)

## Privilege Escalation

- [Privileged Container](./red-team/privileged-container/README.md)
- [Cluster-Admin Binding](./red-team/cluster-admin-binding/README.md)
- [Writable hostPath Mount](./red-team/writable-hostpath/README.md)
- [Access Cloud Resources](./red-team/access-cloud-resources/README.md)

## Defense Evasion

- [Clear Container Logs](./red-team/clear-container-logs/README.md)
- [Delete Kubernetes Events](./red-team/delete-kubernetes-events/README.md)
- [Pod / Container Name Similarity](./red-team/pod-container-name/README.md)
- [Connect from Proxy](./red-team/connect-from-proxy/README.md)

## Credential Access

- [List Kubernetes Secrets](./red-team/list-kubernetes-secrets/README.md)
- [Mount Service Principal](./red-team/mount-service-principal/README.md)
- [Container Service Account](./red-team/container-service-account/README.md)
- [Application Credentials in Configuration Files](./red-team/application-credentials/README.md)
- [Access Managed Identity Credentials](./red-team/access-managed-identity/README.md)
- [Malicious Admission Controller](./red-team/malicious-admission-controller/README.md)

## Discovery

- [Access Kubernetes API Server](./red-team/access-kubernetes-api/README.md)
- [Access Kubelet API](./red-team/access-kubelet-api/README.md)
- [Network Mapping](./red-team/network-mapping/README.md)
- [Exposed Sensitive Interfaces](./red-team/exposed-sensitive-interfaces/README.md)
- [Instance Metadata API](./red-team/instance-metadata-api/README.md)

## Lateral Movement

- [Access Cloud Resources](./red-team/access-cloud-resources/README.md)
- [Container Service Account](./red-team/container-service-account/README.md)
- [Cluster Internal Networking](./red-team/cluster-internal-networking/README.md)
- [Application Credentials in Configuration Files](./red-team/application-credentials/README.md)
- [Writable hostPath Mount](./red-team/writable-hostpath/README.md)
- [CoreDNS Poisoning](./red-team/coredns-poisoning/README.md)
- [ARP Poisoning and IP Spoofing](./red-team/arp-poisoning/README.md)

## Collection

- [Images from a Private Registry](./red-team/images-from-registry/README.md)
- [Collecting Data from Pod](./red-team/collecting-data/README.md)

## Impact

- [Data Destruction](./red-team/data-destruction/README.md)
- [Resource Hijacking](./red-team/resource-hijacking/README.md)
- [Denial of Service](./red-team/denial-of-service/README.md)