#!/bin/bash -ex
# Copyright The Notary Project Authors.
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This file include the script to generate testing certificates for CRL testing.
# The generated files are:
# - certchain_with_crl.pem: the fullchain file that includes the leaf 
#   certificate with CRL, intermediate certificate with invalid OCSP and valid 
#   CRL, and the root certificate.
# - leaf.crl: the CRL file that includes the revoked leaf certificate.
# - leaf.key: the private key of the leaf certificate.
# - leaf_revoked.crl: the CRL file that includes the revoked leaf certificate.
# - intermediate.crl: the CRL file that includes the intermediate certificate.
# - intermediate_revoked.crl: the CRL file that includes the revoked intermediate
# - root.crt: the root certificate.
#
# Note: The script will not run in the pipeline, but we need to keep it for
# future maintenance because generating those test certificates with CRL is not
# easy.

# Create root CA configuration file
cat > root.cnf <<EOF
[ req ]
default_bits       = 2048
prompt             = no
distinguished_name = root_distinguished_name
x509_extensions    = v3_ca

[ root_distinguished_name ]
C  = US
ST = State
L  = City
O  = Organization
OU = OrgUnit
CN = RootCA

[ ca ]
default_ca = CA_default

[ CA_default ]
dir               = ./demoCA
new_certs_dir     = \$dir/newcerts
database          = \$dir/index.txt
serial            = \$dir/serial
private_key       = ./root.key
certificate       = ./root.crt
default_md        = sha256
policy            = policy_any
x509_extensions   = usr_cert
copy_extensions   = copy
default_days      = 36500
default_crl_days  = 36500
crlnumber         = \$dir/crlnumber
crl_extensions    = crl_ext

[ policy_any ]
countryName             = optional
stateOrProvinceName     = optional
localityName            = optional
organizationName        = optional
organizationalUnitName  = optional
commonName              = supplied

[ v3_ca ]
basicConstraints       = critical,CA:TRUE
keyUsage               = critical,keyCertSign,cRLSign
subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid:always,issuer:always

[ crl_ext ]
authorityKeyIdentifier = keyid:always
EOF

# Set up OpenSSL CA directory structure
mkdir -p demoCA/newcerts
touch demoCA/index.txt
echo '1002' > demoCA/serial
echo '1002' > demoCA/crlnumber

# Generate root private key
openssl genrsa -out root.key 2048

# Generate self-signed root certificate with extensions
openssl req -x509 -new -key root.key -sha256 -days 36500 -out root.crt \
  -config root.cnf -extensions v3_ca

# Update intermediate.cnf to include [ca] and [CA_default] sections
cat > intermediate.cnf <<EOF
[ req ]
default_bits       = 2048
prompt             = no
distinguished_name = intermediate_distinguished_name
x509_extensions    = v3_intermediate_ca

[ intermediate_distinguished_name ]
C  = US
ST = State
L  = City
O  = Organization
OU = OrgUnit
CN = IntermediateCA

[ ca ]
default_ca = CA_default

[ CA_default ]
dir               = ./intermediateCA
new_certs_dir     = \$dir/newcerts
database          = \$dir/index.txt
serial            = \$dir/serial
private_key       = ./intermediate.key
certificate       = ./intermediate.crt
default_md        = sha256
policy            = policy_any
x509_extensions   = usr_cert
copy_extensions   = copy
default_days      = 36500
default_crl_days  = 36500
crlnumber         = \$dir/crlnumber
crl_extensions    = crl_ext

[ policy_any ]
countryName             = optional
stateOrProvinceName     = optional
localityName            = optional
organizationName        = optional
organizationalUnitName  = optional
commonName              = supplied

[ v3_intermediate_ca ]
basicConstraints       = critical,CA:TRUE,pathlen:0
keyUsage               = critical,keyCertSign,cRLSign
subjectKeyIdentifier   = hash
authorityKeyIdentifier = keyid:always,issuer:always
crlDistributionPoints  = URI:http://localhost:10086/intermediate.crl
authorityInfoAccess    = OCSP;URI:http://localhost.test/ocsp

[ crl_ext ]
authorityKeyIdentifier = keyid:always
EOF

# Set up OpenSSL CA directory structure for intermediate CA
mkdir -p intermediateCA/newcerts
touch intermediateCA/index.txt
echo '1000' > intermediateCA/serial
echo '1000' > intermediateCA/crlnumber

# Generate intermediate private key
openssl genrsa -out intermediate.key 2048

# Generate intermediate CSR
openssl req -new -key intermediate.key -out intermediate.csr -config intermediate.cnf

# Sign intermediate certificate with root CA
openssl ca -config root.cnf -in intermediate.csr -out intermediate.crt -batch -extensions v3_intermediate_ca -extfile intermediate.cnf -notext

# Update leaf.cnf to remove OCSP server
cat > leaf.cnf <<EOF
[ req ]
default_bits       = 2048
prompt             = no
distinguished_name = req_distinguished_name
req_extensions     = v3_req

[ req_distinguished_name ]
C  = US
ST = State
L  = City
O  = Organization
OU = OrgUnit
CN = LeafCert

[ v3_req ]
basicConstraints = critical,CA:FALSE
keyUsage = critical,digitalSignature
crlDistributionPoints = URI:http://localhost:10086/leaf.crl
EOF

# Generate leaf private key
openssl genrsa -out leaf.key 2048

# Generate leaf certificate signing request (CSR)
openssl req -new -key leaf.key -out leaf.csr -config leaf.cnf

# Sign leaf certificate with intermediate CA
openssl ca -config intermediate.cnf -in leaf.csr -out leaf.crt -batch -extensions v3_req -extfile leaf.cnf -notext

# Generate intermediate CRL using root.cnf (before revocation)
openssl ca -config root.cnf -gencrl -out intermediate.crl

# Convert root CRL to DER format
openssl crl -in intermediate.crl -outform der -out intermediate.crl

# Revoke intermediate certificate using root CA
openssl ca -config root.cnf -revoke intermediate.crt

# Generate intermediate CRL including revoked intermediate certificate
openssl ca -config root.cnf -gencrl -out intermediate_revoked.crl

# Convert intermediate CRL to DER format
openssl crl -in intermediate_revoked.crl -outform der -out intermediate_revoked.crl

# Generate leaf CRL
openssl ca -config intermediate.cnf -gencrl -out leaf.crl

# Convert leaf CRL to DER format
openssl crl -in leaf.crl -outform der -out leaf.crl

# Revoke leaf certificate
openssl ca -config intermediate.cnf -revoke leaf.crt

# Generate leaf CRL including revoked leaf certificate
openssl ca -config intermediate.cnf -gencrl -out leaf_revoked.crl

# Convert leaf CRL to DER format
openssl crl -in leaf_revoked.crl -outform der -out leaf_revoked.crl

# merge leaf cert and root cert to create fullchain file
cat leaf.crt intermediate.crt root.crt > certchain_with_crl.pem

rm -rf leaf.csr leaf.crt leaf.cnf root.srl root.cnf root.key root.crl demoCA intermediate.csr intermediate.cnf intermediate.key intermediate.crt intermediateCA
