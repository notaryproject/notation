#!/bin/bash -e
# this script called by ../run.sh
#
# DockerHub has image pulling limit and request rate limit. Please improve your
# subscription level if you confronts those issues.
#
# Please export $DOCKER_USERNAME and $DOCKER_PASSWORD to use this script.

# check required environment variable
if [ -z "$DOCKER_USERNAME" ] | [ -z "$DOCKER_PASSWORD" ]; then
    echo "\$DOCKER_USERNAME or \$DOCKER_PASSWORD is not set"
    exit 1
fi

# set environment variables for E2E testing
export NOTATION_E2E_REGISTRY_HOST=docker.io/$DOCKER_USERNAME
export NOTATION_E2E_REGISTRY_USERNAME=$DOCKER_USERNAME
export NOTATION_E2E_REGISTRY_PASSWORD=$DOCKER_PASSWORD

function setup_registry {
    echo "use $NOTATION_E2E_REGISTRY_HOST"
}

function cleanup_registry {
    echo "cleaning dockerhub"
    # get token
    # reference: https://docs.docker.com/docker-hub/api/latest/#tag/authentication
    HUB_TOKEN=$(curl -s -H "Content-Type: application/json" -X POST -d "{\"username\": \"$DOCKER_USERNAME\", \"password\": \"$DOCKER_PASSWORD\"}" https://hub.docker.com/v2/users/login/ | jq -r .token)

    for (( page=1;;page++ )); do
        # page query the repositorys' name
        resp=`curl -s -X GET \
            -H "Accept: application/json" \
            -H "Authorization: JWT $HUB_TOKEN" \
            "https://hub.docker.com/v2/repositories/$DOCKER_USERNAME/?page_size=100&&page=$page"`

        # check the last page
        if [[ "$resp" == *"object not found"* ]]; then
            break
        fi

        # parse json and extract e2e repoName
        e2eRepos=(`echo $resp | jq -r '.results|.[]|.name' | grep 'e2e-'`)
        echo "repositories: ${e2eRepos[@]}"

        for repoName in "${e2eRepos[@]}"
        do
            # run delete
            curl -X DELETE \
                -H "Accept: application/json" \
                -H "Authorization: JWT $HUB_TOKEN" \
                https://hub.docker.com/v2/repositories/$DOCKER_USERNAME/$repoName/ && \
                echo "$repoName deleted."
        done
    done
}