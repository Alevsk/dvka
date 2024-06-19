# Terrascan: Static code analyzer for Infrastructure as Code

## Quick Start

Detect compliance and security violations across Infrastructure as Code to mitigate risk before provisioning cloud native infrastructure.

1. Installation.

    Install on your system

    ```bash
    # Download binary
    wget https://github.com/tenable/terrascan/releases/download/v1.18.3/terrascan_1.18.3_Linux_x86_64.tar.gz
    tar -xzf terrascan_1.18.3_Linux_x86_64.tar.gz
    ./terrascan --help
    # Install via brew
    brew install terrascan
    # Initialize
    terrascan init
    ```

    **OR**

    Run via `docker container`

    ```bash
    docker run tenable/terrascan --help
    ```

2. Run `terrascan` command and get familiar with it.

    ```bash
    terrascan
                                             
    Terrascan

    Detect compliance and security violations across Infrastructure as Code to mitigate risk before provisioning cloud native infrastructure.
    For more information, please visit https://runterrascan.io/

    Usage:
    terrascan [command]

    Available Commands:
    completion  Generate the autocompletion script for the specified shell
    help        Help about any command
    init        Initializes Terrascan and clones policies from the Terrascan GitHub repository.
    scan        Detect compliance and security violations across Infrastructure as Code.
    server      Run Terrascan as an API server
    version     Terrascan version

    Flags:
    -c, --config-path string      config file path
    -h, --help                    help for terrascan
    -l, --log-level string        log level (debug, info, warn, error, panic, fatal) (default "info")
        --log-output-dir string   directory path to write the log and output files
    -x, --log-type string         log output type (console, json) (default "console")
    -o, --output string           output type (human, json, yaml, xml, junit-xml, sarif, github-sarif) (default "human")
        --temp-dir string         temporary directory path to download remote repository,module and templates

    Use "terrascan [command] --help" for more information about a command.
    ```

3. Use `terrascan` to scan for vulnerabilities in the `minio` application

    ```bash
    terrascan scan -i k8s minio/
    # run using docker
    docker run --rm -v $(pwd)/minio:/workspace --workdir /workspace tenable/terrascan scan -i k8s -f /workspace/minio-standalone-deployment.yaml
    ```

## Resources

- <https://github.com/tenable/terrascan>
