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


mkdir -p bbr-binary
export BBR_VERSION=1.2.2
curl -L -o bbr-binary/bbr https://s3.amazonaws.com/bosh-dependencies/bbr-$BBR_VERSION

export BBR_SHA256=829160a61a44629a2626b578668777074c7badd75a9b5dab536defdbdd84b17a
export BBR_BINARY_PATH="${PWD}/bbr-binary/bbr"

echo "${BBR_SHA256} ${BBR_BINARY_PATH}" | sha256sum -c -

chmod +x "${BBR_BINARY_PATH}"

cp "${BBR_BINARY_PATH}" /usr/local/bin/bbr

DOCKER_CERTS="$(bosh int /tmp/local-bosh/director/bosh-director.yml --path /instance_groups/0/properties/docker_cpi/docker/tls)"
export DOCKER_CERTS
DOCKER_HOST="$(bosh int /tmp/local-bosh/director/bosh-director.yml --path /instance_groups/name=bosh/properties/docker_cpi/docker/host)"
export DOCKER_HOST

apt-get update
apt-get install -y mysql-client
apt-get install -y postgresql-client
apt-get install -y jq

RDS_MYSQL_EXTERNAL_DB_HOST="$(jq -r .aws_mysql_endpoint database-metadata/metadata | cut -d':' -f1)"
RDS_POSTGRES_EXTERNAL_DB_HOST="$(jq -r .aws_postgres_endpoint database-metadata/metadata | cut -d':' -f1)"
GCP_MYSQL_EXTERNAL_DB_HOST="$(jq -r .gcp_mysql_endpoint database-metadata/metadata)"
GCP_POSTGRES_EXTERNAL_DB_HOST="$(jq -r .gcp_postgres_endpoint database-metadata/metadata)"
GCP_MYSQL_EXTERNAL_DB_CA="$(jq -r .mysql_ca_cert gcp-ssl-config/gcp_mysql.yml)"
GCP_MYSQL_EXTERNAL_DB_CLIENT_CERTIFICATE="$(jq -r .mysql_client_cert gcp-ssl-config/gcp_mysql.yml)"
GCP_MYSQL_EXTERNAL_DB_CLIENT_PRIVATE_KEY="$(jq -r .mysql_client_key gcp-ssl-config/gcp_mysql.yml)"
GCP_POSTGRES_EXTERNAL_DB_CA="$(jq -r .postgres_ca_cert gcp-ssl-config/gcp_postgres.yml)"
GCP_POSTGRES_EXTERNAL_DB_CLIENT_CERTIFICATE="$(jq -r .postgres_client_cert gcp-ssl-config/gcp_postgres.yml)"
GCP_POSTGRES_EXTERNAL_DB_CLIENT_PRIVATE_KEY="$(jq -r .postgres_client_key gcp-ssl-config/gcp_postgres.yml)"

export RDS_MYSQL_EXTERNAL_DB_HOST
export RDS_POSTGRES_EXTERNAL_DB_HOST
export GCP_MYSQL_EXTERNAL_DB_HOST
export GCP_POSTGRES_EXTERNAL_DB_HOST
export GCP_MYSQL_EXTERNAL_DB_CA
export GCP_MYSQL_EXTERNAL_DB_CLIENT_CERTIFICATE
export GCP_MYSQL_EXTERNAL_DB_CLIENT_PRIVATE_KEY
export GCP_POSTGRES_EXTERNAL_DB_CA
export GCP_POSTGRES_EXTERNAL_DB_CLIENT_CERTIFICATE
export GCP_POSTGRES_EXTERNAL_DB_CLIENT_PRIVATE_KEY

cd bosh-src
scripts/test-brats
