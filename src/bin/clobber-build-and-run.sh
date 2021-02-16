#!/usr/bin/env bash

set -e

BIN="$(cd "$(dirname "$0")" ; pwd)"
SRC="$(dirname "${BIN}")"
PROJECT="$(dirname "${SRC}")"

declare -a FLAGS_INHERIT
source "${BIN}/verbose.sh"

if [[ ".$1" = '.--help' ]]
then
    echo "Usage: $(basename "$0") [ -v [ -v ] ] [ --tee <file> ] [ --skip-build ] [ --dev ]" >&2
    echo "       $(basename "$0") --help" >&2
    exit 0
fi

if [[ ".$1" = '.--tee' ]]
then
    exec > >(tee "$2") 2>&1
    shift 2
fi

DO_BUILD='true'
if [[ ".$1" = '.--skip-build' ]]
then
  DO_BUILD='false'
  shift
fi

DO_CLOBBER='true'
if [[ ".$1" = '.--no-clobber' ]]
then
  DO_CLOBBER='false'
fi

: ${AXON_SERVER_PORT=8024}
: ${API_SERVER_PORT=8181}
: ${ENSEMBLE_NAME=example}
"${BIN}/create-local-settings.sh"

source "${PROJECT}/src/etc/settings-local.sh"

function waitForServerReady() {
    local URL="$1"
    local N="$2"
    if [[ -z "${N}" ]]
    then
        N=120
    fi
    while [[ "${N}" -gt 0 ]]
    do
        N=$[$N - 1]
        sleep 1
        if curl -sS "${URL}" >/dev/null 2>&1
        then
            break
        fi
    done
}

(
    cd "${PROJECT}"

    src/bin/generate-root-key-pair.sh "${FLAGS_INHERIT[@]}"
    src/bin/generate-module-for-trusted-keys.sh

    if "${DO_BUILD}"
    then
        # Build server executables from Go sources
        "${BIN}/nix-build.sh"

        # Build docker images for proxy
        docker build -t "${DOCKER_REPOSITORY}/proxy:${IMAGE_VERSION}" src/proxy

        # Build docker image for Swagger UI
        docker build -t "${DOCKER_REPOSITORY}/grpc-swagger" src/swagger
    fi

    (
        cd src/docker
        docker-compose -p "${ENSEMBLE_NAME}" rm --stop --force || true
    )

    if "${DO_CLOBBER}"
    then
      docker volume rm -f "${ENSEMBLE_NAME}_axon-data"
      docker volume rm -f "${ENSEMBLE_NAME}_axon-eventdata"
      docker volume rm -f "${ENSEMBLE_NAME}_elastic-search-data"
    fi

    src/docker/docker-compose-up.sh "${FLAGS_INHERIT[@]}" "$@"
)