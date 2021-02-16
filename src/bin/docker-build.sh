#!/bin/bash

BIN="$(cd "$(dirname "$0")" ; pwd)"
SRC="$(dirname "${BIN}")"
PROJECT="$(dirname "${SRC}")"

source "${BIN}/verbose.sh"

"${BIN}/create-local-settings.sh"

source "${SRC}/etc/settings-local.sh"

if [ -f "${PROJECT}/target/bin/example" ]
then
  docker build --tag "${DOCKER_REPOSITORY}/config-manager:${IMAGE_VERSION}" \
    -f "${PROJECT}/src/docker/Dockerfile-config-manager" "${PROJECT}/target"
else
  info "WARNING: Skipped build of docker image!"
fi
