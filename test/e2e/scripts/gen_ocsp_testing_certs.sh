#!/bin/bash -ex

# Create root configuration with CA and OCSP extensions
cat > root_ocsp.cnf <<EOF
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

[ v3_ca ]
basicConstraints       = critical,CA:TRUE
keyUsage               = critical,keyCertSign,cRLSign
subjectKeyIdentifier   = hash

[ v3_ocsp ]
extendedKeyUsage       = critical,OCSPSigning
keyUsage               = critical,digitalSignature
subjectKeyIdentifier   = hash
EOF

# Set up OpenSSL CA directory structure for root
mkdir -p demoCA/newcerts
touch demoCA/index.txt
echo '1000' > demoCA/serial

# Generate root private key and self-signed certificate with CA extensions
openssl genrsa -out root.key 2048
openssl req -x509 -new -key root.key -sha256 -days 36500 -out root.crt \
  -config root_ocsp.cnf -extensions v3_ca

# Create OCSP responder configuration
cat > ocsp.cnf <<EOF
[ ca ]
default_ca = CA_default

[ CA_default ]
dir               = demoCA
database          = demoCA/index.txt
new_certs_dir     = demoCA/newcerts
certificate       = root.crt
serial            = demoCA/serial
private_key       = root.key
default_days      = 36500
default_md        = sha256
policy            = policy_strict

[ policy_strict ]
countryName             = match
stateOrProvinceName     = match
organizationName        = match
organizationalUnitName  = optional
commonName              = supplied
emailAddress            = optional

[ req ]
default_bits       = 2048
prompt             = no
distinguished_name = ocsp_distinguished_name
req_extensions     = v3_ocsp

[ ocsp_distinguished_name ]
C  = US
ST = State
L  = City
O  = Organization
OU = OrgUnit
CN = OCSPResponder

[ v3_ocsp ]
extendedKeyUsage       = critical,OCSPSigning
keyUsage               = critical,digitalSignature
subjectKeyIdentifier   = hash
EOF

# Generate OCSP private key and CSR, then sign it with the root certificate
openssl genrsa -out ocsp.key 2048
openssl req -new -key ocsp.key -out ocsp.csr -config ocsp.cnf
openssl x509 -req -in ocsp.csr -CA root.crt -CAkey root.key -CAcreateserial \
  -out ocsp.crt -days 36500 -extfile ocsp.cnf -extensions v3_ocsp

# Create leaf configuration with OCSP URL
cat > leaf_ocsp.cnf <<EOF
[ req ]
default_bits = 2048
prompt = no
distinguished_name = leaf_distinguished_name
req_extensions = v3_req

[ leaf_distinguished_name ]
C = US
ST = State
L = City
O = Organization
OU = OrgUnit
CN = LeafCert

[ v3_req ]
basicConstraints       = critical,CA:FALSE
keyUsage               = critical,digitalSignature
authorityInfoAccess    = OCSP;URI:http://localhost:10087
subjectKeyIdentifier   = hash
EOF

# Generate leaf key and CSR then sign directly with the root certificate instead of an intermediate
openssl genrsa -out leaf.key 2048
openssl req -new -key leaf.key -out leaf.csr -config leaf_ocsp.cnf
openssl x509 -req -in leaf.csr -CA root.crt -CAkey root.key -CAcreateserial \
  -out leaf.crt -days 36500 -extfile leaf_ocsp.cnf -extensions v3_req

# Cleanup and final message
rm -f *.csr root_occrt.srl
echo "OCSP testing certificates generated successfully."

