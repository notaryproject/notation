# notation inspect

## Description

Use `notation inspect` command to inspect all the signatures associated with signed artifact in a human readable format.

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

## Outline

```text
Inspect all signatures associated with the signed artifact.

Usage:
    notation inspect [flags] <reference>
  
Flags:
   -h, --help              help for describing the signature
   -o, --output json       output on command line sets the output to json
   -p, --password string   password for registry operations (default to $NOTATION_PASSWORD if not specified)
       --plain-http        registry access via plain HTTP
   -u, --username string   username for registry operations (default to $NOTATION_USERNAME if not specified)
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
    │   │   ├── SHA1 fingerprint: E8C15B4C98AD91E051EE5AF5F524A8729050B2A
    │   │   │   ├── issued to: wabbit-com Software
    │   │   │   ├── issued by: wabbit-com Software Root Certificate Authority
    │   │   │   └── expiry: Sun Jul 06 20:50:17 2025
    │   │   ├── SHA1 fingerprint: 5DCC2147712B3C555B1C96CFCC00215403TF044D
    │   │   │   ├── issued to: wabbit-com Software Code Signing PCA
    │   │   │   ├── issued by: wabbit-com Software Root Certificate Authority
    │   │   │   └── expiry: Sun Jul 06 20:50:17 2025
    │   │   └── SHA1 fingerprint: 1GYA3107712B3C886B1C96AAEC89984914DC0A5A
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
        │   ├── SHA1 fingerprint: 68C15B4C98AD91E051EE5AF5F524A8729040B1D
        │   │   ├── issued to: wabbit-com Software
        │   │   ├── issued by: wabbit-com Software Root Certificate Authority
        │   │   └── expiry: Sun Jul 06 20:50:17 2025
        │   ├── SHA1 fingerprint: 4ACC2147712B3C555B1C96CFCC00215403TE011C
        │   │   ├── issued to: wabbit-com Software Code Signing PCA 2010
        │   │   ├── issued by: wabbit-com Software Root Certificate Authority
        │   │   └── expiry: Sun Jul 06 20:50:17 2025
        │   └── SHA1 fingerprint: A4YA1205512B3C886B1C96AAEC89984914DC012A
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
    │   │   ├── SHA1 fingerprint: E8C15B4C98AD91E051EE5AF5F524A8729050B2A
    │   │   │   ├── issued to: wabbit-com Software
    │   │   │   ├── issued by: wabbit-com Software Root Certificate Authority
    │   │   │   └── expiry: Sun Jul 06 20:50:17 2025
    │   │   ├── SHA1 fingerprint: 5DCC2147712B3C555B1C96CFCC00215403TF044D
    │   │   │   ├── issued to: wabbit-com Software Code Signing PCA
    │   │   │   ├── issued by: wabbit-com Software Root Certificate Authority
    │   │   │   └── expiry: Sun Jul 06 20:50:17 2025
    │   │   └── SHA1 fingerprint: 1GYA3107712B3C886B1C96AAEC89984914DC0A5A
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
        │   ├── SHA1 fingerprint: 68C15B4C98AD91E051EE5AF5F524A8729040B1D
        │   │   ├── issued to: wabbit-com Software
        │   │   ├── issued by: wabbit-com Software Root Certificate Authority
        │   │   └── expiry: Sun Jul 06 20:50:17 2025
        │   ├── SHA1 fingerprint: 4ACC2147712B3C555B1C96CFCC00215403TE011C
        │   │   ├── issued to: wabbit-com Software Code Signing PCA
        │   │   ├── issued by: wabbit-com Software Root Certificate Authority
        │   │   └── expiry: Sun Jul 06 20:50:17 2025
        │   └── SHA1 fingerprint: A4YA1205512B3C886B1C96AAEC89984914DC012A
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
          "SHA1Fingerprint": "E8C15B4C98AD91E051EE5AF5F524A8729050B2A",
          "issuedTo": "wabbit-com Software",
          "issuedBy": "wabbit-com Software Root Certificate Authority",
          "expiry": "2025-07-06T20:50:17Z"
        },
        {
          "SHA1Fingerprint": "5DCC2147712B3C555B1C96CFCC00215403TF044D",
          "issuedTo": "wabbit-com Software Code Signing PCA",
          "issuedBy": "wabbit-com Software Root Certificate Authority",
          "expiry": "2025-07-06T20:50:17Z"
        },
        {
          "SHA1Fingerprint": "1GYA3107712B3C886B1C96AAEC89984914DC0A5A",
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
          "SHA1Fingerprint": "68C15B4C98AD91E051EE5AF5F524A8729040B1D",
          "issuedTo": "wabbit-com Software",
          "issuedBy": "wabbit-com Software Root Certificate Authority",
          "expiry": "2025-07-06T20:50:17Z"
        },
        {
          "SHA1Fingerprint": "4ACC2147712B3C555B1C96CFCC00215403TE011C",
          "issuedTo": "wabbit-com Software Code Signing PCA",
          "issuedBy": "wabbit-com Software Root Certificate Authority",
          "expiry": "2025-07-06T20:50:17Z"
        },
        {
          "SHA1Fingerprint": "A4YA1205512B3C886B1C96AAEC89984914DC012A",
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
