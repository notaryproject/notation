#!/bin/bash -e
# this script called by ../run.sh
#
# Usage
#   ./run.sh zot <notation-binary-path> [old-notation-binary-path]

REG_HOST=localhost
REG_PORT=5000
ZOT_CONTAINER_NAME=zot

# set environment variables for E2E testing
export NOTATION_E2E_REGISTRY_HOST=$REG_HOST:$REG_PORT
export NOTATION_E2E_REGISTRY_USERNAME=testuser
export NOTATION_E2E_REGISTRY_PASSWORD=testpassword

function setup_registry {
    # start Zot
    docker run -d -p $REG_PORT:$REG_PORT -it --name $ZOT_CONTAINER_NAME \
        --mount type=bind,source=`pwd`/testdata/registry/zot/,target=/etc/zot \
        --rm ghcr.io/project-zot/zot-minimal-linux-amd64:latest
    # make sure that Zot is ready
    sleep 1
}

function cleanup_registry {
    docker container stop $ZOT_CONTAINER_NAME 1>/dev/null && echo "Zot stopped"
}