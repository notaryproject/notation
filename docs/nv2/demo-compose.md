# Notary Sandbox with Docker Compose

Launch a notary sandbox with:

```shell
docker-compose up -d --build
docker-compose exec sandbox /bin/bash
```

Setup sandbox:

```shell
ln -s /certs/reg-ca/ca.crt /usr/local/share/ca-certificates/reg-ca.crt
update-ca-certificates
mkdir -p ~/.docker
```

Setup nv2 folder with signing key and cli:

```shell
mkdir nv2-demo
cd nv2-demo/
openssl req \
  -x509 \
  -sha256 \
  -nodes \
  -newkey rsa:2048 \
  -days 365 \
  -subj "/CN=registry.wabbit-networks.io/O=wabbit-networks inc/C=US/ST=Washington/L=Seattle" \
  -addext "subjectAltName=DNS:registry.wabbit-networks.io" \
  -keyout ./wabbit-networks.key \
  -out ./wabbit-networks.crt
alias docker="docker nv2"
```

Build and sign an image:

```shell
export REPO=registry.wabbit-networks.io/net-monitor
export IMAGE=${REPO}:v1
docker build \
  -t $IMAGE \
  https://github.com/wabbit-networks/net-monitor.git#main
docker notary --enabled
docker notary sign \
  --key ./wabbit-networks.key \
  --cert ./wabbit-networks.crt \
  $IMAGE
docker push $IMAGE
```

Inspect image on registry:

```shell
oras discover $IMAGE --output tree
```

Pull signed image (fails without signature):

```shell
docker pull $IMAGE
```

Trust wabbit key and pull again:

```shell
cat >~/.docker/nv2.json <<EOF
{
  "enabled": true,
  "verificationCerts": [
    "/root/nv2-demo/wabbit-networks.crt"
  ],
  "insecureRegistries": [
  ]
}
EOF
docker pull $IMAGE
```

Create a sample SBoM artifact:

```shell
echo '{"version": "0.0.0.0", "artifact": "net-monitor:v1", "contents": "good"}' > sbom.json
oras push $REPO \
  --artifact-type application/x.example.sbom.v0 \
  --artifact-reference $IMAGE \
  --export-manifest sbom-manifest.json \
  ./sbom.json:application/tar
oras discover $IMAGE --output tree
```

Sign the SBoM:

```shell
DIGEST=$(oras discover \
    --artifact-type application/x.example.sbom.v0 \
    --output json \
    $IMAGE | jq -r .references[0].digest)
nv2 sign \
  -m x509 \
  -k wabbit-networks.key \
  -c wabbit-networks.crt \
  --push \
  --push-reference oci://${REPO}@${DIGEST} \
  file:sbom-manifest.json
oras discover ${REPO}@${DIGEST} --output tree
```

Exit the sandbox:

```shell
exit
```

Cleanup:

```shell
docker-compose down
```

Also cleanup volumes:

```shell
docker-compose down -v
```

## Development environment

To build with changes to nv2 and oras, checkout oras within the current directory so that `./oras` exist:

```shell
git clone https://github.com/deislabs/oras.git
cd oras; git checkout reference-types; cd ..
```

Perform any changes to this nv2 repo or the oras repo.
Then start with the dev compose file merged in:

```shell
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d
docker-compose exec sandbox /bin/bash
```

Then update the installed binaries with code from these repos:

```shell
rebuild.sh
```
