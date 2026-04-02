# Example Dockerfile showing how an attacker backdoors a legitimate base image.
# Base image is a well-known, trusted image. The attacker adds a malicious layer
# that runs silently alongside the legitimate process.
#
# Build: docker build -t attacker-registry.example.com/nginx:1.25-alpine .
# Push:  docker push attacker-registry.example.com/nginx:1.25-alpine
#
# The resulting image looks identical to the original from the outside.
# Image scan tools that only check the base digest will miss the added layer.

FROM nginx:1.25-alpine

# Add a "backdoor" script that runs at container startup.
# In a real attack this would be a reverse shell, a crypto miner,
# or a credential harvester - all hidden inside a COPY layer.
COPY backdoor.sh /docker-entrypoint.d/00-backdoor.sh
RUN chmod +x /docker-entrypoint.d/00-backdoor.sh

# The legitimate entrypoint and CMD are preserved so the container
# behaves normally and does not raise operator suspicion.
