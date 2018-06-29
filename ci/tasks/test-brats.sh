#!/usr/bin/env bash

set -eu

script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
src_dir="${script_dir}/../../.."

"${src_dir}/bosh-src/ci/docker/main-bosh-docker/start-bosh.sh"

source /tmp/local-bosh/director/env

bosh int /tmp/local-bosh/director/creds.yml --path /jumpbox_ssh/private_key > /tmp/jumpbox_ssh_key.pem
chmod 400 /tmp/jumpbox_ssh_key.pem

export BOSH_DIRECTOR_IP="10.245.0.3"

BOSH_BINARY_PATH=$(which bosh)
export BOSH_BINARY_PATH
export BOSH_RELEASE="${PWD}/bosh-src/src/spec/assets/dummy-release.tgz"
export BOSH_DIRECTOR_RELEASE_PATH="${PWD}/bosh-release"
DNS_RELEASE_PATH="$(realpath "$(find "${PWD}"/bosh-dns-release -maxdepth 1 -path '*.tgz')")"
export DNS_RELEASE_PATH
CANDIDATE_STEMCELL_TARBALL_PATH="$(realpath "${src_dir}"/stemcell/*.tgz)"
export CANDIDATE_STEMCELL_TARBALL_PATH
export BOSH_DEPLOYMENT_PATH="/usr/local/bosh-deployment"
export BOSH_DNS_ADDON_OPS_FILE_PATH="${BOSH_DEPLOYMENT_PATH}/experimental/dns-addon-with-api-certificates.yml"

export OUTER_BOSH_ENV_PATH="/tmp/local-bosh/director/env"

DOCKER_CERTS="$(bosh int /tmp/local-bosh/director/bosh-director.yml --path /instance_groups/0/properties/docker_cpi/docker/tls)"
export DOCKER_CERTS
DOCKER_HOST="$(bosh int /tmp/local-bosh/director/bosh-director.yml --path /instance_groups/name=bosh/properties/docker_cpi/docker/host)"
export DOCKER_HOST

cd bosh-src
scripts/test-brats
