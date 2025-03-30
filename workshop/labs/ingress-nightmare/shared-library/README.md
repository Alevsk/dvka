# Ingress Nightmare Shared Libraries

Simple shared libraries that demonstrate the CVE-2025-1974 vulnerability in ingress-nginx.

1. Shared library that prints "hello world".

    ```bash
    gcc -fPIC -shared -o hello_engine.so hello_engine.c -lcrypto
    ```

1. Reverse shell shared library that starts a reverse shell.

    ```bash
    gcc -fPIC -shared -o reverse_shell.so reverse_shell.c -lcrypto
    ```

