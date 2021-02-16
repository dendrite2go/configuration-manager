#!/usr/bin/false

DOCKER_REPOSITORY='dendrite2go'
NIX_STORE_VOLUME="${USER}-nix-store"
EXAMPLE_IMAGE_VERSION='0.0.1-SNAPSHOT'
UI_SERVER_PORT='3000'
API_SERVER_PORT='8181'
AXON_SERVER_PORT='8024'
AXON_VERSION='4.3.1'
ELASTIC_SEARCH_VERSION='7.6.1'
ROOT_PRIVATE_KEY='/Users/jeroen/.ssh/id_rsa'
##ROOT_PRIVATE_KEY='data/secure/id_rsa'
SIGN_PRIVATE_KEY='data/secure/id_rsa'
ADDITIONAL_TRUSTED_KEYS=()

EXTRA_VOLUMES="
      -
        type: bind
        source: ${PROJECT}
        target: ${PROJECT}"
