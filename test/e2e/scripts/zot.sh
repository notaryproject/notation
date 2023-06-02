#!/bin/bash -e
# this script called by ../run.sh
#
# Usage
#   ./run.sh zot <notation-binary-path> [old-notation-binary-path]

source ./scripts/tls.sh

REG_HOST=localhost
REG_PORT=5000
REG_PORT_WITHOUT_REFERRERS_API=5001
ZOT_CONTAINER_NAME=notation-e2e-registry
NGINX_CONTAINER_NAME=nginx
DOCKER_NETWORK=notation-e2e

# set required environment variables for E2E testing
export NOTATION_E2E_REGISTRY_HOST_WITH_REFERRERS_API="$REG_HOST:$REG_PORT"
export NOTATION_E2E_REGISTRY_HOST_WITHOUT_REFERRERS_API="$REG_HOST:$REG_PORT_WITHOUT_REFERRERS_API"
export NOTATION_E2E_REGISTRY_USERNAME=testuser
export NOTATION_E2E_REGISTRY_PASSWORD=testpassword
export NOTATION_E2E_REGISTRY_HOST_FOR_DOMAIN="$DOMAIN"

function setup_registry {
    docker network create "$DOCKER_NETWORK"
    # start Zot
    docker run -d -p "$REG_PORT:$REG_PORT" -it \
        --name "$ZOT_CONTAINER_NAME" \
        --network "$DOCKER_NETWORK" \
        --mount type=bind,source="$(pwd)/testdata/registry/zot/",target=/etc/zot \
        --rm ghcr.io/project-zot/zot-minimal-linux-amd64:latest

    # start Nginx for TLS testing and emulate a registry that doesn't support 
    # referrers api.
    docker run -d -p 80:80 -p 443:443 -p "$REG_PORT_WITHOUT_REFERRERS_API:$REG_PORT_WITHOUT_REFERRERS_API" \
        --network "$DOCKER_NETWORK" \
        --mount type=bind,source="$(pwd)/testdata/nginx/",target=/etc/nginx \
        --name "$NGINX_CONTAINER_NAME" \
        --rm nginx:latest

    if [ "$GITHUB_ACTIONS" == "true" ]; then
        setup_system_tls
    fi

    # make sure that Zot is ready
    sleep 1
}

function cleanup_registry {
    docker container stop "$ZOT_CONTAINER_NAME" 1>/dev/null && echo "Zot stopped"
    docker container stop "$NGINX_CONTAINER_NAME" 1>/dev/null && echo "Nginx stopped"
    docker network rm "$DOCKER_NETWORK" 1>/dev/null && echo "Docker network $DOCKER_NETWORK removed"

    if [ "$GITHUB_ACTIONS" == "true" ]; then
        clean_up_system_tls
    fi
}
