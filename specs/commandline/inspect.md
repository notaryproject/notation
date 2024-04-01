# notation inspect

## Description

Use `notation inspect` command to inspect all the signatures associated with artifacts stored in OCI compliant registries in a human readable format.

Upon successful execution, both the digest of the signed artifact and the digests of signatures manifest along with their properties associated with the signed artifact are printed in the following format:

```shell
<registry>/<repository>@<digest>
└── application/vnd.cncf.notary.signature
    ├── <digest of signature manifest>
    │   ├── <signature algorithm>
    │   ├── <signed attributes>
    │   ├── <user defined attributes>
    │   ├── <unsigned attributes>
    │   ├── <certificates>
    │   └── <signed artifact>
    └── <digest of signature manifest>
        ├── <signature algorithm>
        ├── <signed attributes>
        ├── <unsigned attributes>
        ├── <certificates>
        └── <signed artifact>
```

> [!NOTE]
> This command is for inspecting signatures associated with OCI artifacts only. Use `notation blob inspect` command for inspecting signatures associated with arbitrary blobs.

## Outline

```text
Inspect all signatures associated with a signed OCI artifact.

Usage:
    notation inspect [flags] <reference>

Flags:
      --allow-referrers-api   [Experimental] use the Referrers API to inspect signatures, if not supported (returns 404), fallback to the Referrers tag schema
  -d, --debug                 debug mode
  -h, --help                  help for inspect
      --insecure-registry     use HTTP protocol while connecting to registries. Should be used only for testing
      --max-signatures int    maximum number of signatures to evaluate or examine (default 100)
  -o, --output string         output format, options: 'json', 'text' (default "text")
  -p, --password string       password for registry operations (default to $NOTATION_PASSWORD if not specified)
  -u, --username string       username for registry operations (default to $NOTATION_USERNAME if not specified)
  -v, --verbose               verbose mode
```

## Usage

### Display the details of all the listed signatures and its associated certificate properties of the signed container image


```text
notation inspect [flags] <registry>/<repository>@<digest>
```

## Inspect signatures on the supplied OCI artifact identified by the digest

```shell
# Prerequisites: Signatures are stored in a registry referencing the signed OCI artifact
notation inspect localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da1ac484efe37a5380ee9088f7ace2efcde9
```

An example output:
```shell
Inspecting all signatures for signed artifact
localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac4efe37a5380ee9088f7ace2efcde9
└── application/vnd.cncf.notary.signature
    ├── sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
    │   ├── signature algorithm: RSASSA-PSS-SHA-256
    │   ├── signed attributes
    │   │   ├── content type: application/vnd.cncf.notary.payload.v1+json
    │   │   ├── signing scheme: notary.default.x509
    │   │   ├── signing time: Fri Jun 23 22:04:01 2023
    │   │   ├── expiry: Sat Jun 29 22:04:01 2024
    │   │   └── io.cncf.notary.verificationPlugin: com.example.nv2plugin    //extended attributes
    │   ├── user defined attributes
    │   │   └── io.wabbit-networks.buildId: 123                             //user defined metadata
    │   ├── unsigned attributes
    │   │   ├── io.cncf.notary.timestampSignature: <Base64(TimeStampToken)> //TSA response
    │   │   └── io.cncf.notary.signingAgent: notation/1.0.0                 //client version
    │   ├── certificates
    │   │   ├── SHA256 fingerprint: E8C15B4C98AD91E051EE5AF5F524A8729050B2A
    │   │   │   ├── issued to: wabbit-com Software
    │   │   │   ├── issued by: wabbit-com Software Root Certificate Authority
    │   │   │   └── expiry: Sun Jul 06 20:50:17 2025
    │   │   ├── SHA256 fingerprint: 4b9fa61d5aed0fabbc7cb8fe2efd049da57957ed44f2b98f7863ce18effd3b89
    │   │   │   ├── issued to: wabbit-com Software Code Signing PCA
    │   │   │   ├── issued by: wabbit-com Software Root Certificate Authority
    │   │   │   └── expiry: Sun Jul 06 20:50:17 2025
    │   │   └── SHA256 fingerprint: ea3939548ad0c0a86f164ab8b97858854238c797f30bddeba6cb28688f3f6536
    │   │       ├── issued to: wabbit-com Software Root Certificate Authority
    │   │       ├── issued by: wabbit-com Software Root Certificate Authority
    │   │       └── expiry: Sat Jun 23 22:04:01 2035
    │   └── signed artifact                                                 //descriptor of signed artifact
    │       ├── media type: application/vnd.oci.image.manifest.v1+json
    │       ├── digest: sha256:b94d27b9934d3e08a52e52d7da7dabfac48437a5380ee9088f7ace2efcde9
    │       └── size: 16724
    └── sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
        ├── signature algorithm: RSASSA-PSS-SHA-256
        ├── signed attributes
        │   ├── content type: application/vnd.cncf.notary.payload.v1+json
        │   ├── signing scheme: notary.signingAuthority.x509
        │   ├── signing time: Fri Jun 23 22:04:01 2023
        │   ├── expiry: Sat Jun 29 22:04:01 2024
        │   └── io.cncf.notary.verificationPlugin: com.example.nv2plugin
        ├── unsigned attributes
        │   ├── io.cncf.notary.timestampSignature: <Base64(TimeStampToken)>
        │   └── io.cncf.notary.signingAgent: notation/1.0.0
        ├── certificates
        │   ├── SHA256 fingerprint: b13a843be16b1f461f08d61c14f3eab7d87c073570da077217541a7eb31c084d
        │   │   ├── issued to: wabbit-com Software
        │   │   ├── issued by: wabbit-com Software Root Certificate Authority
        │   │   └── expiry: Sun Jul 06 20:50:17 2025
        │   ├── SHA256 fingerprint: 4b9fa61d5aed0fabbc7cb8fe2efd049da57957ed44f2b98f7863ce18effd3b89
        │   │   ├── issued to: wabbit-com Software Code Signing PCA 2010
        │   │   ├── issued by: wabbit-com Software Root Certificate Authority
        │   │   └── expiry: Sun Jul 06 20:50:17 2025
        │   └── SHA256 fingerprint: ea3939548ad0c0a86f164ab8b97858854238c797f30bddeba6cb28688f3f6536
        │       ├── issued to: wabbit-com Software Root Certificate Authority
        │       ├── issued by: wabbit-com Software Root Certificate Authority
        │       └── expiry: Sat Jun 23 22:04:01 2035
        └── signed artifact
            ├── media type: application/vnd.oci.image.manifest.v1+json
            ├── digest: sha256:b94d27b9934d3e08a52e52d7da7fac484efe37a5380ee9088f7ace2efcde9
            └── size: 16724
```

## Inspect signatures on an OCI artifact identified by a tag

`Tags` are mutable, but `Digests` uniquely and immutably identify an artifact. If a tag is used to identify a signed artifact, notation resolves the tag to the `digest` first.

```shell
# Prerequisites: Signatures are stored in a registry referencing the signed OCI artifact
notation inspect localhost:5000/net-monitor:v1
```

An example output:
```text
Resolved artifact tag `v1` to digest `sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9` before inspect.
Warning: The resolved digest may not point to the same signed artifact, since tags are mutable.
```

```shell
Inspecting all signatures for signed artifact
localhost:5000/net-monitor@sha256:ca5427b5567d3e06a72e52d7da7dabfac484efe37a5380ee9088f7ace2eaab9
└── application/vnd.cncf.notary.signature
    ├── sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
    │   ├── signature algorithm: RSASSA-PSS-SHA-256
    │   ├── signed attributes
    │   │   ├── content type: application/vnd.cncf.notary.payload.v1+json
    │   │   ├── signing scheme: notary.default.x509
    │   │   ├── signing time: Fri Jun 23 22:04:01 2023
    │   │   ├── expiry: Sat Jun 29 22:04:01 2024
    │   │   └── io.cncf.notary.verificationPlugin: com.example.nv2plugin
    │   ├── user defined attributes
    │   │   └── io.wabbit-networks.buildId: 123
    │   ├── unsigned attributes
    │   │   ├── io.cncf.notary.timestampSignature: <Base64(TimeStampToken)>
    │   │   └── io.cncf.notary.signingAgent: notation/1.0.0
    │   ├── certificates
    │   │   ├── SHA256 fingerprint: b13a843be16b1f461f08d61c14f3eab7d87c073570da077217541a7eb31c084d
    │   │   │   ├── issued to: wabbit-com Software
    │   │   │   ├── issued by: wabbit-com Software Root Certificate Authority
    │   │   │   └── expiry: Sun Jul 06 20:50:17 2025
    │   │   ├── SHA256 fingerprint: 4b9fa61d5aed0fabbc7cb8fe2efd049da57957ed44f2b98f7863ce18effd3b89
    │   │   │   ├── issued to: wabbit-com Software Code Signing PCA
    │   │   │   ├── issued by: wabbit-com Software Root Certificate Authority
    │   │   │   └── expiry: Sun Jul 06 20:50:17 2025
    │   │   └── SHA256 fingerprint: ea3939548ad0c0a86f164ab8b97858854238c797f30bddeba6cb28688f3f6536
    │   │       ├── issued to: wabbit-com Software Root Certificate Authority
    │   │       ├── issued by: wabbit-com Software Root Certificate Authority
    │   │       └── expiry: Sat Jun 23 22:04:01 2035
    │   └── signed artifact
    │       ├── media type: application/vnd.oci.image.manifest.v1+json
    │       ├── digest: sha256:ca5427b5567d3e06a72e52d7da7dabfac484efe37a5380ee9088f7ace2eaab9
    │       └── size: 16724
    └── sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
        ├── signature algorithm: RSASSA-PSS-SHA-256
        ├── signed attributes
        │   ├── content type: application/vnd.cncf.notary.payload.v1+json
        │   ├── signing scheme: notary.signingAuthority.x509
        │   ├── signing time: Fri Jun 23 22:04:01 2023
        │   ├── expiry: Sat Jun 29 22:04:01 2024
        │   └── io.cncf.notary.verificationPlugin: com.example.nv2plugin
        ├── unsigned attributes
        │   ├── io.cncf.notary.timestampSignature: <Base64(TimeStampToken)>
        │   └── io.cncf.notary.signingAgent: notation/1.0.0
        ├── certificates
        │   ├── SHA256 fingerprint: b13a843be16b1f461f08d61c14f3eab7d87c073570da077217541a7eb31c084d
        │   │   ├── issued to: wabbit-com Software
        │   │   ├── issued by: wabbit-com Software Root Certificate Authority
        │   │   └── expiry: Sun Jul 06 20:50:17 2025
        │   ├── SHA256 fingerprint: 4b9fa61d5aed0fabbc7cb8fe2efd049da57957ed44f2b98f7863ce18effd3b89
        │   │   ├── issued to: wabbit-com Software Code Signing PCA
        │   │   ├── issued by: wabbit-com Software Root Certificate Authority
        │   │   └── expiry: Sun Jul 06 20:50:17 2025
        │   └── SHA256 fingerprint: ea3939548ad0c0a86f164ab8b97858854238c797f30bddeba6cb28688f3f6536
        │       ├── issued to: wabbit-com Software Root Certificate Authority
        │       ├── issued by: wabbit-com Software Root Certificate Authority
        │       └── expiry: Sat Jun 23 22:04:01 2035
        └── signed artifact
            ├── media type: application/vnd.oci.image.manifest.v1+json
            ├── digest: sha256:ca5427b5567d3e06a72e52d7da7dabfac484efe37a5380ee9088f7ace2eaab9
            └── size: 16724
```
## Inspect signatures on the supplied OCI artifact with an example of JSON Output

```shell
notation inspect localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52da7dabfac484efe37a5380ee9088f7ace2efcde9 -o json
```

An example output:
```jsonc
{
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "signatures": [
    {
      "digest": "sha256:73c803930ea3ba1e54bc25c2bdc53edd0284c62ed651fe7b00369da519a33",
      "signatureAlgorithm": "RSASSA-PSS-SHA-256",
      "signedAttributes": {
        "contentType": "application/vnd.cncf.notary.payload.v1+json",
        "signingScheme": "notary.default.x509",
        "signingTime": "2022-02-06T20:50:17Z",
        "expiry": "2023-02-06T20:50:17Z",
        "io.cncf.notary.verificationPlugin": "com.example.nv2plugin"
      },
      "userDefinedAttributes": {
        "io.wabbit-networks.buildId": "123"
      },
      "unsignedAttributes": {
        "io.cncf.notary.timestampSignature": "<Base64(TimeStampToken)>",
        "io.cncf.notary.signingAgent": "notation/1.0.0"
      },
      "certificates": [
        {
          "SHA256Fingerprint": "b13a843be16b1f461f08d61c14f3eab7d87c073570da077217541a7eb31c084d",
          "issuedTo": "wabbit-com Software",
          "issuedBy": "wabbit-com Software Root Certificate Authority",
          "expiry": "2025-07-06T20:50:17Z"
        },
        {
          "SHA256Fingerprint": "4b9fa61d5aed0fabbc7cb8fe2efd049da57957ed44f2b98f7863ce18effd3b89",
          "issuedTo": "wabbit-com Software Code Signing PCA",
          "issuedBy": "wabbit-com Software Root Certificate Authority",
          "expiry": "2025-07-06T20:50:17Z"
        },
        {
          "SHA256Fingerprint": "ea3939548ad0c0a86f164ab8b97858854238c797f30bddeba6cb28688f3f6536",
          "issuedTo": "wabbit-com Software Root Certificate Authority",
          "issuedBy": "wabbit-com Software Root Certificate Authority",
          "expiry": "2035-07-06T20:50:17Z"
        }
      ],
      "signedArtifact": {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "digest": "sha256:73c803930ea3ba1e54bc25c2bdc53edd0284c62ed651fe7b00369519a3c333",
        "size": 16724
      }
    },
    {
      "digest": "sha256:73c803930ea3ba1e54bc25c2bdc53edd0284c62ed651fe7b00369da519a3c333",
      "signatureAlgorithm": "RSASSA-PSS-SHA-256",
      "signedAttributes": {
        "contentType": "application/vnd.cncf.notary.payload.v1+json",
        "signingScheme": "notary.signingAuthority.x509",
        "signingTime": "2022-02-06T20:50:17Z",
        "expiry": "2023-02-06T20:50:17Z",
        "io.cncf.notary.verificationPlugin": "com.example.nv2plugin"
      },
      "unsignedAttributes": {
        "io.cncf.notary.timestampSignature": "<Base64(TimeStampToken)>",
        "io.cncf.notary.signingAgent": "notation/1.0.0"
      },
      "certificates": [
        {
          "SHA256Fingerprint": "b13a843be16b1f461f08d61c14f3eab7d87c073570da077217541a7eb31c084d",
          "issuedTo": "wabbit-com Software",
          "issuedBy": "wabbit-com Software Root Certificate Authority",
          "expiry": "2025-07-06T20:50:17Z"
        },
        {
          "SHA256Fingerprint": "4b9fa61d5aed0fabbc7cb8fe2efd049da57957ed44f2b98f7863ce18effd3b89",
          "issuedTo": "wabbit-com Software Code Signing PCA",
          "issuedBy": "wabbit-com Software Root Certificate Authority",
          "expiry": "2025-07-06T20:50:17Z"
        },
        {
          "SHA256Fingerprint": "ea3939548ad0c0a86f164ab8b97858854238c797f30bddeba6cb28688f3f6536",
          "issuedTo": "wabbit-com Software Root Certificate Authority",
          "issuedBy": "wabbit-com Software Root Certificate Authority",
          "expiry": "2035-07-06T20:50:17Z"
        }
      ],
      "signedArtifact": {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "digest": "sha256:73c803930ea3ba1e54bc25c2bdc53edd0284c62ed651fe7b069da519a3c333",
        "size": 16724
      }
    }
  ]
}
```
