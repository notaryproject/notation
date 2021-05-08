#!/bin/sh

set -ex

: "${REG_CA_KEY:=/certs/private/ca.key}"
: "${REG_CA_CERT:=/certs/shared/ca.crt}"
: "${REG_SAN:=DNS:registry}"

# create key for ca and registry if missing
if [ -n "${REG_CA_KEY}" -a ! -f "${REG_CA_KEY}" ]; then
  mkdir -p "$(dirname "${REG_CA_KEY}")"
  openssl genrsa -out "${REG_CA_KEY}" 4096
fi
if [ -n "${REGISTRY_HTTP_TLS_KEY}" -a ! -f "${REGISTRY_HTTP_TLS_KEY}" ]; then
  mkdir -p "$(dirname "${REGISTRY_HTTP_TLS_KEY}")"
  openssl genrsa -out "${REGISTRY_HTTP_TLS_KEY}" 4096
fi

# regenerate cert for ca on every start, avoids dealing with expiration
mkdir -p "$(dirname "${REG_CA_CERT}")"
openssl req -new -key "${REG_CA_KEY}" \
  -out "${REG_CA_CERT}" \
  -subj '/CN=Registry CA' -x509 -days "3650"

# regenerate cert for registry on every start, use SAN from env
if [ -n "${REGISTRY_HTTP_TLS_CERTIFICATE}" ]; then
  mkdir -p "$(dirname "${REGISTRY_HTTP_TLS_CERTIFICATE}")"
  openssl req -new \
    -key "${REGISTRY_HTTP_TLS_KEY}" \
    -out "${REGISTRY_HTTP_TLS_KEY}.csr" \
    -subj '/CN=Registry server'
  echo "[ x509_exts ]" > "${REGISTRY_HTTP_TLS_KEY}.cnf" 
  echo "subjectAltName = ${REG_SAN}" >> "${REGISTRY_HTTP_TLS_KEY}.cnf"
  openssl x509 -req \
    -in "${REGISTRY_HTTP_TLS_KEY}.csr" \
    -CA "${REG_CA_CERT}" \
    -CAkey "${REG_CA_KEY}" \
    -CAcreateserial \
    -out "${REGISTRY_HTTP_TLS_CERTIFICATE}" \
    -days "365" \
    -extfile "${REGISTRY_HTTP_TLS_KEY}.cnf" \
    -extensions x509_exts
fi

exec registry "$@"