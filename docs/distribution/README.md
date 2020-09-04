# OCI Distribution

There is an ongoing discussion to modify OCI index objects to include a config property (see [proposal](https://github.com/notaryproject/nv2/pull/10)). Such a property implies a relationship between the manifests referenced in the index and its config object. For a referenced manifest, one can create a reverse lookup to the index config. With this ability, manifest signatures that are pushed to a registry as index config objects can be stored and retrieved as manifest referrer metadata.

## Table of contents

1. [Manifest Referrer](#manifest-referrer)
2. [Storing Signatures](#storing-signatures)
3. [Retrieving Signatures](#retrieving-signatures)
4. [Implementation](#implementation)
5. [Prototype](#prototype)

## Manifest Referrer

A manifest referrer is any registry artifact that has an immutatable reference to a manifest. An OCI index is a referrer to each manifest it references. Currently, the OCI image spec does not include a config property for an OCI index and there is no reverse lookup of referrers in docker distribution.\
A modified OCI index with a config property that references a collection of manifests allows us to associate a "type" to the referrer-referenced relationship, where the type is the `mediaType` of the index config object, such as `application/vnd.cncf.notary.config.v2+jwt`.

## Storing Signatures

The proposal is to implement a referrer metadata store for manifests that is essentially a reverse-lookup, by `mediaType`, to referrer config objects. For example, when an OCI index is pushed, if it references a config object of media type `application/vnd.cncf.notary.config.v2+jwt`, a link to the config object is recorded in the referrer metadata store of each referenced manifest.

> See [Artifacts submitted to a registry](https://github.com/notaryproject/nv2/blob/276abe450feec8f3b54f0774b7dfb47f36670cc5/docs/distribution/README.md#artifacts-submitted-to-a-registry) for each digest reference used in this example.

### Put an OCI index by digest, linking a signature to a collection of manifests

### Request

`PUT https://localhost:5000/v2/net-monitor/manifests/sha256:222ibbf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cb222i`

```json
{
  "schemaVersion": 3,
  "mediaType": "application/vnd.oci.image.index.v1+json",
  "config": {
    "mediaType": "application/vnd.cncf.notary.config.v2+jwt",
    "digest": "sha256:222cb130c152895905abe66279dd9feaa68091ba55619f5b900f2ebed38b222c",
    "size": 1906
  },
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:111ma2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49a111m",
      "size": 7023,
      "platform": {
        "architecture": "ppc64le",
        "os": "linux"
      }
    }
  ]
}
```

PUT index would result in the creation of a link between the index config object `sha256:222cb130c152895905abe66279dd9feaa68091ba55619f5b900f2ebed38b222c`and the `net-monitor` manifest `sha256:111ma2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49a111m`, of type `application/vnd.cncf.notary.config.v2+jwt`.

## Retrieving Signatures

Signatures are just referrer metadata to a registry. Referrer metdata for a manifest can be retrieved filtered by their types.

### Get a list of paginated signatures for a manifest by tag

### Request

`GET http://localhost:5000/v2/net-monitor/manifests/v1.0/referrer-metadata?media-type=application/vnd.cncf.notary.config.v2+jwt`

> note: the mediaType filter will move to a request header

### Response

The request will return the two signatures (**wabbit-networks** & **acme-rockets**)

> note: the response should be a complete oci-descriptor for each result and root reference

```json
{
    "tag": "v1.0",
    "@nextLink": "{opaqueUrl}",
    "referrerMetadata": [
        "sha256:222ibbf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cb222i",
        "sha256:333ic0c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75c333i"
    ]
}
```

### Get a list of paginated signatures for a manifest by digest

### Request

`GET http://localhost:5000/v2/net-monitor/manifests/sha256:111ma2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49a111m/referrer-metadata?media-type=application/vnd.cncf.notary.config.v2+jwt`

> note: the mediaType filter will move to a request header

### Response

```json
{
    "digest": "sha256:111ma2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49a111m",
    "@nextLink": "{opaqueUrl}",
    "referrerMetadata": [
        "sha256:222ibbf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cb222i",
        "sha256:333ic0c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75c333i"
    ]
}
```

## Implementation

Let's consider an example implementation for docker distribution, backed by file storage. Say that an image already exists in the registry:

- repository: `net-monitor`
- digest: `sha256:111ma2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49a111m`
- tag: `v1.0`

The storage layout we're most concerned with at the moment is the repository store where the manifest link file exists. It's shown below:

```bash
<root>
└── v2
    └── repositories
        └── net-monitor
            └── _manifests
                └── revisions
                    └── sha256
                        └── 111ma2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49a111m
                            └── link
```

Now we push a signature blob and an OCI index that contains a config property referencing it:

- index digest: `sha256:222ibbf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cb222i`
- index json:
    ```json
    {
        "schemaVersion": 3,
        "mediaType": "application/vnd.oci.image.index.v1+json",
        "config": {
            "mediaType": "application/vnd.cncf.notary.config.v2+jwt",
            "digest": "sha256:222cb130c152895905abe66279dd9feaa68091ba55619f5b900f2ebed38b222c",
            "size": 1906
        },
        "manifests": [
            {
              "mediaType": "application/vnd.oci.image.manifest.v1+json",
              "digest": "sha256:111ma2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49a111m",
              "size": 7023,
              "platform": {
                "architecture": "ppc64le",
                "os": "linux"
              }
            }
        ]
    }
    ```

On `PUT` index to the repository `net-monitor`, the index appears as a manifest revision as usual. Additionally, a link is added to the referrer metadata store of the manifest. The manifests storage layout would look as follows:

```
<root>
└── v2
    └── repositories
        └── net-monitor
            └── _manifests
                └── revisions
                    └── sha256
                        ├── 111ma2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49a111m
                        │   ├── link
                        │   └── referrerMetadata
                        │       └── application/vnd.cncf.notary.config.v2+jwt
                        │           └── sha256
                        │               └── 222cb130c152895905abe66279dd9feaa68091ba55619f5b900f2ebed38b222c
                        │                   └── link
                        └── 222ibbf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cb222i
                            └── link
```

Let's add another signature for the manifest `sha256:111ma2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49a111m`:

- index digest: `sha256:333ic0c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75c333i`
- index json:
  ```json
  {
    "schemaVersion": 3,
    "mediaType": "application/vnd.oci.image.index.v1+json",
    "config": {
      "mediaType": "application/vnd.cncf.notary.config.v2+jwt",
      "digest": "sha256:333cc44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b785c333c",
      "size": 1906
    },
    "manifests": [
      {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "digest": "sha256:111ma2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49a111m",
        "size": 7023,
        "platform": {
          "architecture": "ppc64le",
          "os": "linux"
        }
      }
    ]
  }
  ```

The manifest storage layout would look as follows on `PUT` index:

```
<root>
└── v2
    └── repositories
        └── net-monitor
            └── _manifests
                └── revisions
                    └── sha256
                        ├── 111ma2d22ae5ef400769fa51c84717264cd1520ac8d93dc071374c1be49a111m
                        │   ├── link
                        │   └── referrerMetadata
                        │       └── application/vnd.cncf.notary.config.v2+jwt
                        │           └── sha256
                        │               ├── 222cb130c152895905abe66279dd9feaa68091ba55619f5b900f2ebed38b222c
                        │               │    └── link
                        │               └── 333cc44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b785c333c
                        │                   └── link
                        ├── 222ibbf80b44ce6be8234e6ff90a1ac34acbeb826903b02cfa0da11c82cb222i
                        │   └── link
                        └── 333ic0c33ebc4a74a0a554c86ac2b28ddf3454a5ad9cf90ea8cea9f9e75c333i
                            └── link
```

## Prototype

Available here: https://github.com/notaryproject/distribution/tree/prototype-1

The following steps illustrate how signatures can be stored and retrieved from a registry.

### Prerequisites

- Local registry prototype instance
- [docker-generate](https://github.com/shizhMSFT/docker-generate)
- [nv2](https://github.com/notaryproject/nv2)
- `curl`
- `jq`
- `python3`

### Push an image to your registry

```shell
# Local registry
regIp="127.0.0.1" && \
  regPort="5000" && \
  registry="$regIp:$regPort" && \
  repo="busybox" && \
  tag="latest" && \
  image="$repo:$tag" && \
  reference="$registry/$image"

# Pull image from docker hub and push to local registry
docker pull $image && \
  docker tag $image $reference && \
  docker push $reference
```

### Generate image manifest and sign it

```shell
# Generate self-signed certificates
openssl req \
  -x509 \
  -sha256 \
  -nodes \
  -newkey rsa:2048 \
  -days 365 \
  -subj "/CN=$regIp/O=example inc/C=IN/ST=Haryana/L=Gurgaon" \
  -addext "subjectAltName=IP:$regIp" \
  -keyout example.key \
  -out example.crt

# Generate image manifest
manifestFile="manifest-to-sign.json" && \
  docker generate manifest $image > $manifestFile

# Sign manifest
signatureFile="manifest-signature.jwt" && \
  nv2 sign --method x509 \
    -k example.key \
    -c example.crt \
    -r $reference \
    -o $signatureFile \
    file:$manifestFile
```

### Obtain manifest and signature digests

```shell
manifestDigest="sha256:$(sha256sum $manifestFile | cut -d " " -f 1)" && \
  signatureDigest="sha256:$(sha256sum $signatureFile | cut -d " " -f 1)"
```

### Create an OCI index file referencing the manifest that was signed and its signature as config

```shell
indexFile="index.json" && \
  indexMediaType="application/vnd.oci.image.index.v1+json" && \
  configMediaType="application/vnd.cncf.notary.config.v2+jwt" && \
  signatureFileSize=`wc -c < $signatureFile` && \
  manifestMediaType="$(cat $manifestFile | jq -r '.mediaType')" && \
  manifestFileSize=`wc -c < $manifestFile`

cat <<EOF > $indexFile
{
  "schemaVersion": 3,
  "mediaType": "$indexMediaType",
  "config": {
    "mediaType": "$configMediaType",
    "digest": "$signatureDigest",
    "size": $signatureFileSize
  },
  "manifests": [
    {
      "mediaType": "$manifestMediaType",
      "digest": "$manifestDigest",
      "size": $manifestFileSize
    }
  ]
}
EOF
```

### Obtain index digest

```shell
indexDigest="sha256:$(sha256sum $indexFile | cut -d " " -f 1)"
```

### Push signature and index

```shell
# Initiate blob upload and obtain PUT location
configPutLocation=`curl -I -X POST -s http://$registry/v2/$repo/blobs/uploads/ | grep "Location: " | sed -e "s/Location: //;s/$/\&digest=$signatureDigest/;s/\r//"`

# Push signature blob
curl -X PUT -H "Content-Type: application/octet-stream" --data-binary @"$signatureFile" $configPutLocation

# Push index
curl -X PUT --data-binary @"$indexFile" -H "Content-Type: $indexMediaType" "http://$registry/v2/$repo/manifests/$indexDigest"
```

### Retrieve signatures of a manifest as referrer metadata

```shell
# URL encode index config media type
metadataMediaType=`python3 -c "import urllib.parse, sys; print(urllib.parse.quote(sys.argv[1]))" $configMediaType`

# Retrieve referrer metadata
curl -s "http://$registry/v2/$repo/manifests/$manifestDigest/referrer-metadata?media-type=$metadataMediaType" | jq
```

### Verify signature

```shell
# Retrieve first signature and store it locally
metadataDigest=`curl -s "http://$registry/v2/$repo/manifests/$manifestDigest/referrer-metadata?media-type=$metadataMediaType" | jq -r '.referrerMetadata[0]'` && \
  retrievedMetadataFile="retrieved-signature.jwt" && \
  curl -s http://$registry/v2/$repo/blobs/$metadataDigest > $retrievedMetadataFile

# Verify signature
nv2 verify \
  -f $retrievedMetadataFile \
  -c example.crt \
  file:$manifestFile
```
