# KubeLinter: Check Kubernetes YAML files and Helm charts

## Quick Start

KubeLinter is a static analysis tool that checks Kubernetes YAML files and Helm charts to ensure the applications represented in them adhere to best practices.

1. Installation.

    Install on your system

    ```bash
    # Download binary
    wget https://github.com/stackrox/kube-linter/releases/download/v0.6.5/kube-linter-linux.tar.gz
    tar -xzf kube-linter-linux.tar.gz
    ./kube-linter
    # Using Go
    go install golang.stackrox.io/kube-linter/cmd/kube-linter@latest
    # Using Homebrew for macOS or LinuxBrew for Linux
    brew install kube-linter
    ```

2. Run `kube-linter` command and get familiar with it.

    ```bash
    kube-linter                                                     12:36:45
    Usage:
    kube-linter [command]

    Available Commands:
    checks      View more information on lint checks
    completion  Generate the autocompletion script for the specified shell
    help        Help about any command
    lint        Lint Kubernetes YAML files and Helm charts
    templates   View more information on check templates
    version     Print version and exit

    Flags:
    -h, --help         help for kube-linter
        --with-color   Force color output (default true)

    Use "kube-linter [command] --help" for more information about a command.
    ```

3. Use `kube-linter` to scan for vulnerabilities in the `wordpress` application

    ```bash
    kube-linter lint wordpress/*.yaml
    # report in json format
    kube-linter lint wordpress/*.yaml --format json
    # filter specific fields using jq
    kube-linter lint wordpress/*.yaml --format json | jq -r '[.Reports[] | { "Check": .Check, "Service": .Object.K8sObject.Name, "Type": .Object.K8sObject.GroupVersionKind.Kind, "Message": .Diagnostic.Message, "Remediation": .Remediation }]'
    ```

4. USe `kube-linter` to scan for vulnerabilities in the wordpress `helm` application

    ```bash
    kube-linter lint wordpress-helm-chart/
    ```

## Resources

- <https://github.com/stackrox/kube-linter>
- <https://kubernetes.io/docs/tutorials/stateful-application/mysql-wordpress-persistent-volume/>
