#!/bin/bash

BIN="$(cd "$(dirname "$0")" ; pwd)"
SRC="$(dirname "${BIN}")"
PROJECT="$(dirname "${SRC}")"

source "${BIN}/verbose.sh"
source "${SRC}/etc/settings-local.sh"

if [[ -z "${SIGN_PRIVATE_KEY}" ]]
then
  SIGN_PRIVATE_KEY="${ROOT_PRIVATE_KEY}"
fi

(
  cd "${PROJECT}" || exit 1
  log "DIR=[$(pwd)]"

  ROOT_PUBLIC_KEY="$(cat "${ROOT_PRIVATE_KEY}.pub")"
  ROOT_KEY_NAME="$(cat "${ROOT_PRIVATE_KEY}.pub" | cut -d ' ' -f 3)"
  SIGN_KEY_NAME="$(cat "${SIGN_PRIVATE_KEY}.pub" | cut -d ' ' -f 3)"

  (
    echo ">>> Manager: ${ROOT_KEY_NAME}"
    cat "${ROOT_PRIVATE_KEY}"
    if [[ ".${SIGN_PRIVATE_KEY}" != ".${ROOT_PRIVATE_KEY}" ]]
    then
      echo '>>> Trusted:'
      cat "${SIGN_PRIVATE_KEY}.pub"
    fi

    echo ">>> Identity Provider: ${SIGN_KEY_NAME}"
    cat "${SIGN_PRIVATE_KEY}"

    echo ">>> Secrets"
    cat "${SRC}/etc/secrets-local.yaml" \
      | docker run --rm -i karlkfi/yq -r '.users | to_entries[] | .key + " " + .value.secret' \
      | while read USER_ID PASSWORD_ENCRYPTED
        do
          log ">>> ${USER_ID}: ${PASSWORD_ENCRYPTED}"
          echo "${USER_ID}=${PASSWORD_ENCRYPTED}"
        done

    if [[ -f ""${SRC}/etc/application-local.yaml"" ]]
    then
      echo ">>> Properties"
      cat "${SRC}/etc/application-local.yaml" \
        | docker run --rm -i karlkfi/yq -r --stream \
            '. | select(length > 1) | .[0][0] + (reduce .[0][1:][] as $item ("" ; . + "." + ($item | tostring))) + "=" + .[1]'
    fi

    echo '>>> End'
  ) | docker run -i dendrite2go/configmanager
)