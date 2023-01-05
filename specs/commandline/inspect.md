# notation inspect

## Description

Use `notation inspect` command to inspect all the signatures associated with signed artifact in a human readable format.

Upon successful execution, both the digest of the signed artifact and the digests of signatures manifest along with their properties associated with the signed artifact are printed in the following format:

```shell
<registry>/<repository>@<digest>
└──application/vnd.cncf.notary.signature
  ├──<digest_of_signature_manifest>
    ├──<signingAlgorithm>
    ├──<signedAttributes>
    ├──<userDefinedAttributes>
    ├──<unsignedAttributes>
    ├──<certificates>
    └──<payload>
  ├──<digest_of_signature_manifest>
    ├──<signingAlgorithm>
    ├──<signedAttributes>
    ├──<unsignedAttributes>
    ├──<certificates>
    └──<payload>
```

## Outline

```text
Inspect all signatures with the signed artifact.

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
notation inspect localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
```

An example output:
```shell
localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
└──application/vnd.cncf.notary.signature
    ├── sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
      ├──signingAlgorithm: RSASSA-PSS-SHA-256
      ├──signedAttributes
            ├──contentType: application/vnd.cncf.notary.payload.v1+json
            ├──signingScheme: notary.default.x509
            ├──signingTime: 2022-04-06T07:01:20Z
            ├──expiry: 2022-10-06T07:01:20Z
            ├──io.cncf.notary.verificationPlugin: com.example.nv2plugin     //extended attributes to support plugins
      ├──userDefinedAttributes
            ├──io.wabbit-networks.buildId: 123                              //user defined payload annotations.
      ├──unsignedAttributes
            ├──io.cncf.notary.timestampSignature: <Base64(TimeStampToken)>  //how TSA response (time stamp token) is represented.
            ├──io.cncf.notary.signingAgent: notation/1.0.0                  //provides the identifier of a client (e.g. Notation) that produced the signature
      ├──certificates
            ├──SHA1Fingerprint: 2f1cc5b8455381cdefac83b4bd305b789cc9c16e
              ├──issuedTo: Microsoft Root Certificate Authority 2010
              ├──issuedBy: Microsoft Root Certificate Authority 2010
              ├──expires: Sat Jun 23 22:04:01 2035
            ├──SHA1Fingerprint: 8BFE3107712B3C886B1C96AAEC89984914DC9B6B
              ├──issuedTo: Microsoft Code Signing PCA 2010
              ├──issuedBy: Microsoft Root Certificate Authority 2010
              ├──expires: Sun Jul 06 20:50:17 2025
      └──payload                                                           //descriptor of the target artifact manifest that is signed.
            ├──mediaType: application/vnd.oci.image.manifest.v1+json
            ├──digest: sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
            └──size: 16724
      
    ├── sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
      ├──signingAlgorithm: RSASSA-PSS-SHA-256
      ├──signedAttributes
            ├──contentType: application/vnd.cncf.notary.payload.v1+json
            ├──signingScheme: notary.signingAuthority.x509
            ├──signingTime: 2022-04-06T07:01:20Z
            ├──expiry: 2022-10-06T07:01:20Z
            ├──io.cncf.notary.verificationPlugin: com.example.nv2plugin                  
      ├──unsignedAttributes
            ├──io.cncf.notary.timestampSignature: <Base64(TimeStampToken)>
            ├──io.cncf.notary.signingAgent: notation/1.0.0                  
      ├──certificates
            ├──SHA1Fingerprint: 2f1rr5b8455381frdajc83b4bd305b743cc9513u
              ├──issuedTo: Microsoft Root Certificate Authority 2010
              ├──issuedBy: Microsoft Root Certificate Authority 2010
              ├──expires: Sat Jun 23 22:04:01 2035
            ├──SHA1Fingerprint: 8BFE3107712B3C886B1C96AAEC89984914DC9B6B
              ├──issuedTo: Microsoft Code Signing PCA 2010
              ├──issuedBy: Microsoft Root Certificate Authority 2010
              ├──expires: Sun Jul 06 20:50:17 2025
      └──payload 
            ├──mediaType: application/vnd.oci.image.manifest.v1+json
            ├──digest: sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9
            └──size: 16724
```

## Usage signatures on an OCI artifact identified by a tag

```text
`Tags` are mutable, but `Digests` uniquely and immutably identify an artifact. If a tag is used to identify a signed artifact, notation resolves the tag to the `digest` first.
```

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
localhost:5000/net-monitor@sha256:ca5427b5567d3e06a72e52d7da7dabfac484efe37a5380ee9088f7ace2eaab9
└──application/vnd.cncf.notary.signature
    ├── sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
      ├──signingAlgorithm: RSASSA-PSS-SHA-256
      ├──signedAttributes
            ├──contentType: application/vnd.cncf.notary.payload.v1+json
            ├──signingScheme: notary.default.x509
            ├──signingTime: 2022-04-06T07:01:20Z
            ├──expiry: 2022-10-06T07:01:20Z
            ├──io.cncf.notary.verificationPlugin: com.example.nv2plugin
      ├──userDefinedAttributes
            ├──io.wabbit-networks.buildId: 123
      ├──unsignedAttributes
            ├──io.cncf.notary.timestampSignature: <Base64(TimeStampToken)>
            ├──io.cncf.notary.signingAgent: notation/1.0.0
      ├──certificates
            ├──SHA1Fingerprint: 2f1cc5b8455381cdefac83b4bd305b789cc9c16e
              ├──issuedTo: Microsoft Root Certificate Authority 2010
              ├──issuedBy: Microsoft Root Certificate Authority 2010
              ├──expires: Sat Jun 23 22:04:01 2035
            ├──SHA1Fingerprint: 8BFE3107712B3C886B1C96AAEC89984914DC9B6B
              ├──issuedTo: Microsoft Code Signing PCA 2010
              ├──issuedBy: Microsoft Root Certificate Authority 2010
              ├──expires: Sun Jul 06 20:50:17 2025
      └──payload 
            ├──mediaType: application/vnd.oci.image.manifest.v1+json
            ├──digest: sha256:ca5427b5567d3e06a72e52d7da7dabfac484efe37a5380ee9088f7ace2eaab9
            └──size: 16724
      
    ├── sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
      ├──signingAlgorithm: RSASSA-PSS-SHA-256
      ├──signedAttributes
            ├──contentType: application/vnd.cncf.notary.payload.v1+json
            ├──signingScheme: notary.signingAuthority.x509
            ├──signingTime: 2022-04-06T07:01:20Z
            ├──expiry: 2022-10-06T07:01:20Z
            ├──io.cncf.notary.verificationPlugin: com.example.nv2plugin     
      ├──unsignedAttributes
            ├──io.cncf.notary.timestampSignature: <Base64(TimeStampToken)> 
            ├──io.cncf.notary.signingAgent: notation/1.0.0   
      ├──certificates
            ├──SHA1Fingerprint: 2f1rr5b8455381frdajc83b4bd305b743cc9513u
              ├──issuedTo: Microsoft Root Certificate Authority 2010
              ├──issuedBy: Microsoft Root Certificate Authority 2010
              ├──expires: Sat Jun 23 22:04:01 2035
            ├──SHA1Fingerprint: 8BFE3107712B3C886B1C96AAEC89984914DC9B6B
              ├──issuedTo: Microsoft Code Signing PCA 2010
              ├──issuedBy: Microsoft Root Certificate Authority 2010
              ├──expires: Sun Jul 06 20:50:17 2025
      └──payload 
            ├──mediaType: application/vnd.oci.image.manifest.v1+json
            ├──digest: sha256:ca5427b5567d3e06a72e52d7da7dabfac484efe37a5380ee9088f7ace2eaab9
            └──size: 16724
```
## Inspect signatures on the supplied OCI artifact with an example of JSON Output

```shell
notation inspect localhost:5000/net-monitor@sha256:b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9 -o json
```

An example output:
```jsonc
{
  "signatures": [
    {
      "digest": "sha256:73c803930ea3ba1e54bc25c2bdc53edd0284c62ed651fe7b00369da519a3c333",
      "signingAlgorithm": "RSASSA-PSS-SHA-256",
      "signedAttributes": {
        "contentType": "application/vnd.cncf.notary.payload.v1+json",
        "signingScheme": "notary.default.x509",
        "signingTime": "2022-04-06T07:01:20Z",
        "expiry": "2022-10-06T07:01:20Z",
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
          "SHA1Fingerprint": "8BFE3107712B3C886B1C96AAEC89984914DC9B6B",
          "issuedTo": "Microsoft Root Certificate Authority 2010",
          "issuedBy": "Microsoft Root Certificate Authority 2010",
          "expires": "Sun Jul 06 20:50:17 2025"
        },
        {
          "SHA1Fingerprint": "8BFE3107712B3C886B1C96AAEC89984914DC9B6B",
          "issuedTo": "Microsoft Code Signing PCA 2010",
          "issuedBy": "Microsoft Root Certificate Authority 2010",
          "expires": "Sun Jul 06 20:50:17 2025"
        }
      ],
      "payload": {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "digest": "sha256:73c803930ea3ba1e54bc25c2bdc53edd0284c62ed651fe7b00369da519a3c333",
        "size": "16724"
      }
    },
    {
      "digest": "sha256:73c803930ea3ba1e54bc25c2bdc53edd0284c62ed651fe7b00369da519a3c333",
      "signingAlgorithm": "RSASSA-PSS-SHA-256",
      "signedAttributes": {
        "contentType": "application/vnd.cncf.notary.payload.v1+json",
        "signingScheme": " notary.signingAuthority.x509",
        "signingTime": "2022-04-06T07:01:20Z",
        "expiry": "2022-10-06T07:01:20Z",
        "io.cncf.notary.verificationPlugin": "com.example.nv2plugin"
      },
      "unsignedAttributes": {
        "io.cncf.notary.timestampSignature": "<Base64(TimeStampToken)>",
        "io.cncf.notary.signingAgent": "notation/1.0.0"
      },
      "certificates": [
        {
          "SHA1Fingerprint": "8BFE3107712B3C886B1C96AAEC89984914DC9B6B",
          "issuedTo": "Microsoft Code Signing PCA 2010",
          "issuedBy": "Microsoft Root Certificate Authority 2010",
          "expires": "Sun Jul 06 20:50:17 2025"
        },
        {
          "SHA1Fingerprint": "8BFE3107712B3C886B1C96AAEC89984914DC9B6B",
          "issuedTo": "Microsoft Code Signing PCA 2010",
          "issuedBy": "Microsoft Root Certificate Authority 2010",
          "expires": "Sun Jul 06 20:50:17 2025"
        }
      ],
      "payload": {
        "mediaType": "application/vnd.oci.image.manifest.v1+json",
        "digest": "sha256:73c803930ea3ba1e54bc25c2bdc53edd0284c62ed651fe7b00369da519a3c333",
        "size": "16724"
      }
    }
  ]
}
```
