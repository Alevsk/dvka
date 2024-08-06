# SSH server running inside container

SSH server that is running inside a container may be used by attackers. If attackers gain valid credentials to a container, whether by brute force attempts or by other methods (such as phishing), they can use it to get remote access to the container by SSH.

## Resources

- <https://hub.docker.com/r/linuxserver/openssh-server>