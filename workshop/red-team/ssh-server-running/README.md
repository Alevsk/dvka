# SSH Server Running

This document describes how an attacker can use an SSH server running in a container to gain remote access.

## Description

Attackers may run an SSH server in a container to get a remote shell to the container. This can be used to execute commands in the container and to exfiltrate data from the container.

## Resources

- [OpenSSH Server on Docker Hub](https://hub.docker.com/r/linuxserver/openssh-server)
