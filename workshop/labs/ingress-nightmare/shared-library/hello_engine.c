#include <stdio.h>
#include <string.h>
#include <openssl/engine.h>
#include <unistd.h>
#include <sys/types.h>

/**
 * OpenSSL Engine for the Ingress Nightmare Lab
 * 
 * This is a simple demonstration engine that logs information
 * about the process it's running in.
 */

/* Engine identification constants */
static const char *engine_id   = "hello";
static const char *engine_name = "Hello World Test Engine";

/**
 * Engine binding function called when the engine is loaded
 *
 * @param e The engine structure to initialize
 * @param id The engine ID passed during loading
 * @return 1 on success, 0 on failure
 */
static int bind_hello(ENGINE *e, const char *id)
{
    /* Log if there's an ID mismatch, but continue anyway */
    if (id && strcmp(id, engine_id) != 0) {
        fprintf(stderr, "ingress-nightmare lab: Engine invoked with ID: '%s' (expected: '%s')\n", 
                id, engine_id);
    }

    /* Set the engine's internal ID and name */
    if (!ENGINE_set_id(e, engine_id) || 
        !ENGINE_set_name(e, engine_name)) {
        fprintf(stderr, "ingress-nightmare lab: Failed to set engine id or name\n");
        return 0; /* Failure */
    }

    /* Log initialization success */
    printf("ingress-nightmare lab: Hello World engine initialized successfully\n");
    
    /* Log process information for debugging */
    pid_t pid = getpid();
    uid_t uid = getuid();
    printf("ingress-nightmare lab: Process ID: %d, User ID: %d\n", pid, uid);

    return 1; /* Success */
}

/* Required OpenSSL engine implementation macros */
IMPLEMENT_DYNAMIC_BIND_FN(bind_hello)
IMPLEMENT_DYNAMIC_CHECK_FN()
