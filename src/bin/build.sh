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

(
    cd "${PROJECT}"

    src/bin/generate-root-key-pair.sh "${FLAGS_INHERIT[@]}"
    src/bin/generate-module-for-trusted-keys.sh

    # Build server executables from Go sources
    "${BIN}/nix-build.sh"

    # Build docker image
    "${BIN}/docker-build.sh"
)
