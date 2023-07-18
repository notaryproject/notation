#!/bin/bash -e
# Copyright The Notary Project Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# this script called by ../run.sh
#
# Usage
#   ./run.sh zot <notation-binary-path> [old-notation-binary-path]

source ./scripts/tls.sh

REG_HOST=localhost
REG_PORT=5000
ZOT_CONTAINER_NAME=notation-e2e-registry

# set required environment variables for E2E testing
export NOTATION_E2E_REGISTRY_HOST="$REG_HOST:$REG_PORT"
export NOTATION_E2E_REGISTRY_USERNAME=testuser
export NOTATION_E2E_REGISTRY_PASSWORD=testpassword
export NOTATION_E2E_DOMAIN_REGISTRY_HOST="$DOMAIN"

function setup_registry {
    create_docker_network
    # start Zot
    docker run -d -p $REG_PORT:$REG_PORT -it \
        --name $ZOT_CONTAINER_NAME \
        --network $DOCKER_NETWORK \
        --mount type=bind,source="$(pwd)/testdata/registry/zot/",target=/etc/zot \
        --rm ghcr.io/project-zot/zot-minimal-linux-amd64:latest

    if [ "$GITHUB_ACTIONS" == "true" ]; then
        setup_tls
    fi
    # make sure that Zot is ready
    sleep 1
}

function cleanup_registry {
    docker container stop $ZOT_CONTAINER_NAME 1>/dev/null && echo "Zot stopped"
    if [ "$GITHUB_ACTIONS" == "true" ]; then
        clean_up_tls
    fi
    remove_docker_network
}
