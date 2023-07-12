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

#
# Usage
#   For setup:
#   1. source ./scripts/tls.sh
#   2. call create_docker_network
#   3. setup registry with port 5000 in $DOCKER_NETWORK
#   4. call setup_tls reverse proxy
#
#   For clean up:
#   1. call clean_up
#   2. clean up registry
#   3. call remove_docker_network
# 
# Note: this script needs sudo permission to add TLS certificate to system and 
#       add domain registry host.

NGINX_CONTAINER_NAME=nginx
DOMAIN=notation-e2e.registry.io
DOCKER_NETWORK=notation-e2e

function create_docker_network {
    docker network create "$DOCKER_NETWORK"
}

function remove_docker_network {
    docker network rm "$DOCKER_NETWORK"
}

function setup_tls {
    # add domain registry host to /etc/hosts for testing --plain-http feature
    echo "127.0.0.1 $DOMAIN" | sudo tee -a /etc/hosts
    # add TLS certificate to system
    sudo mkdir -p /usr/local/share/ca-certificates/
    sudo cp ./testdata/nginx/notation-e2e.registry.io.crt /usr/local/share/ca-certificates/
    sudo update-ca-certificates

    # start Nginx for TLS 
    docker run -d -p 80:80 -p 443:443 \
        --network "$DOCKER_NETWORK" \
        --mount type=bind,source="$(pwd)/testdata/nginx/",target=/etc/nginx \
        --name "$NGINX_CONTAINER_NAME" \
        --rm nginx:latest
}

function clean_up_tls {
    docker container stop "$NGINX_CONTAINER_NAME" 1>/dev/null && echo "Nginx stopped"
    sudo sed -i "/${NOTATION_E2E_DOMAIN_REGISTRY_HOST}/d" /etc/hosts
    sudo rm /usr/local/share/ca-certificates/notation-e2e.registry.io.crt
    sudo update-ca-certificates
}
