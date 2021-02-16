#!/bin/bash

BIN="$(cd "$(dirname "$0")" ; pwd)"
SRC="$(dirname "${BIN}")"
PROJECT="$(dirname "${SRC}")"

source "${BIN}/verbose.sh"

"${BIN}/create-local-settings.sh"

source "${SRC}/etc/settings-local.sh"

docker build --tag "${DOCKER_REPOSITORY}/archetype-go-axon:latest" "${SRC}/docker"

if [ -f "${PROJECT}/target/bin/example" ]
then
  docker build --tag "${DOCKER_REPOSITORY}/config-manager:latest" \
    -f "${PROJECT}/src/docker/Dockerfile-config-manager" "${PROJECT}/target"
fi
