# Notary v2 Specification

This specification describes the file formats, queries, and processes required to use TUF in registries. Primarily those requirements are that metadata for each of the TUF roles root, snapshot, timestamp, and targets be uploaded to a registry and that this metadata can be queried by users. Details about each of these roles can be found in the [design document](https://github.com/notaryproject/nv2/blob/prototype-tuf/tuf-design.md) and the [TUF specification](https://github.com/theupdateframework/specification/blob/master/tuf-spec.md#the-update-framework-specification).

## Root metadata

Root metadata should be signed using an offline process. It can then be uploaded to the registry by running `nv2 upload root`.

## Signing a manifest

To upload a manifest signed with Notary v2, a developer must:
* Generate and sign targets metadata by running `nv2 sign imagename keyloc --role targetsrolename` where `imagename` is the manifest to be signed, `targetsrolename` is the name of the top-level or delegated targets role the developer is a part of, and `keyloc` is a path to the developer's private key.
* Upload the manifest and targets metadata to the registry by running `nv2 upload imagename`.
* Send the targets metadata to the snapshot process. This may be part of the above command, but the developer should specify a `SNAPSHOT_LOCATION`.

## Configuring delegations

As developers join or leave teams, it may be necessary to update delegations. To do so, the delegator should perform the following steps:
* To add a new delegatee run `nv2 delegate fromrole torole --namespace ns --key keyloc`. This should add `torole` as a delegation from `fromrole` that is trusted for images in the namespace `ns`.
* To remove a delegatee run `nv2 revoke fromrole torole --key keyloc`.
* After performing the above commands as many times as needed, run `nv2 upload fromrole`. This will upload the new `fromrole` metadata. Any new delegated roles should sign and upload their targets metadata as well.


## Snapshot and timestamp process

For each TUF root, there must be a process that generates snapshot and timestamp metadata. This process should be automated and may be hosted by a registry or an organization that uploads to a registry. It must include the following steps:

* Receive a new targets metadata version. A developer who plans to upload new TUF targets metadata must send information about its version number to the snapshot process. The process may perform any verification steps to ensure the developers are authorized to upload a new version.
* Add metadata version to snapshot metadata. The version number of the new metadata must be added to the next version of snapshot metadata.
* Sign snapshot metadata. This may be done using an online key.
* Update timestamp metadata with the new hash of the snapshot metadata.
* Sign timestamp metadata. This may be done using an online key.
* Upload snapshot and timestamp metadata to the registry.

Changes to snapshot metadata may be batched, so multiple targets metadata versions may be added at once. To do so, the process can allow steps 1 and 2 to repeat any number of times before signing the snapshot metadata.

## User download

When installing an image from a registry, a user may run `nv2 download imagename`. This command must download the following from the registry:
* Root metadata. The client should download the most recent root metadata, as well as all previous versions until they reach their currently trusted root version.
* The most recent timestamp metadata.
* The most recent snapshot metadata.
* the most recent top-level targets metadata.
* Relevant delegated targets metadata. If the top-level targets metadata delegated the queried namespace to another targets metadata file, that file should be downloaded as well. Be aware that there may be multiple levels of delegated targets metadata.
* The requested image.

The downloaded metadata files will be used in TUF verification as described in the [TUF specification](https://github.com/theupdateframework/specification/blob/master/tuf-spec.md#the-update-framework-specification).


## Garbage Collection
Once new versions of snapshot, timestamp, or targets have been uploaded, older versions on the registry can safely be garbage collected. This garbage collection should happen after a registry-specified wait time to ensure that no in progress download is interrupted.

Old versions of root should stay on the registry to allow users with older root keys to verify the rotation of root keys.

## Metadata Types

### Signatures

All metadata files will include a signature in the following format.

```
{
  "signed" : ROLE,
  "signatures" : [
    { "keyid" : KEYID,
      "sig" : SIGNATURE }
      , ... ]
}
```

where

* `ROLE` is a dictionary whose "_type" field describes the role type, either `root`, `timestamp`, `snapshot`, or `targets`.
* `KEYID` is the unique identifier of the key signing the `ROLE` object.
* `SIGNATURE` is a hex-encoded signature of the canonical form of the metadata for `ROLE`.

### Keys

Keys will be described in the following format:

```
{
  "keytype" : KEYTYPE,
  "scheme" : SCHEME,
  "keyval" : KEYVAL
}
```

where

* `KEYTYPE` is a string denoting a public key signature system, such as `rsa`, `ed25519` or `ecdsa-sha2-nistp256`.
* `SCHEME` is a string denoting a corresponding signature scheme.  For example: `rsassa-pss-sha256`, `ed25519`, and `ecdsa-sha2-nistp256`.
* `KEYVAL` is a dictionary containing the public portion of the key.

### Root

Root metadata will contain the following in the `ROLE` field:

```
{
  "_type" : "root",
  "spec_version" : SPEC_VERSION,
  "consistent_snapshot": CONSISTENT_SNAPSHOT,
  "version" : VERSION,
  "expires" : EXPIRES,
  "keys" : {
    KEYID : KEY,
    ...
  },
  "roles" : {
    ROLE : {
      "keyids" : [
        KEYID,
        ...
      ] ,
      "threshold" : THRESHOLD
    },
    ...
  }
}
```

where
* `SPEC_VERSION` is a string that contains the version number of the TUF specification. Its format follows the Semantic Versioning 2.0.0 (semver) specification. Metadata is written according to version "spec_version" of the specification, and clients MUST verify that "spec_version" matches the expected version number. Adopters are free to determine what is considered a match (e.g., the version number exactly, or perhaps only the major version number (major.minor.fix).
* `CONSISTENT_SNAPSHOT` is a boolean indicating whether the repository supports consistent snapshots. The [TUF specification](https://github.com/theupdateframework/specification/blob/master/tuf-spec.md#consistent-snapshots--consistent-snapshots) goes into more detail on the consequences of enabling this setting on a repository.
* `VERSION` is an integer that is greater than 0. Clients MUST NOT replace a metadata file with a version number less than the one currently trusted.
* `EXPIRES` is a date-time string indicating when metadata should be considered expired and no longer trusted by clients. Clients MUST NOT trust an expired file.
* `ROLE` is one of "root", "snapshot", "targets", or "timestamp". Each of these roles MUST be specified in the key list.
* `KEYID` is a KEYID, which MUST be correct for the specified KEY. Clients MUST ensure that for any KEYID represented in this key list and in other files, only one unique key has that KEYID.
* `THRESHOLD` is an integer number of role keys whose signatures are required in order to consider a file as being properly signed by that role.

### Snapshot

Snapshot metadata will contain the following in the `ROLE` field:

```
{
  "_type" : "snapshot",
  "spec_version" : SPEC_VERSION,
  "version" : VERSION,
  "expires" : EXPIRES,
  "meta" : METAFILES
}
```

* `SPEC_VERSION`, `VERSION` and `EXPIRES` are the same as is described for the root.json file.
* METAFILES is an object whose format is the following:
```
{
  METATAG : {
    "version" : VERSION,
    ("length" : LENGTH,)
  },
  ...
}
```
* `METATAG` is a string giving the fully qualified reference of the metadata on the repository. For snapshot.json, these are top-level targets metadata and delegated targets metadata.
* `VERSION` is an integer version number as shown in the metadata file at METATAG.
* `LENGTH` is an integer length in bytes of the metadata file at METATAG. It is OPTIONAL and can be omitted to reduce the snapshot metadata file size. In that case the client MUST use a custom download limit for the listed metadata.

### Targets

Targets metadata will contain the following in the `ROLE` field:

```
{
  "_type" : "targets",
  "spec_version" : SPEC_VERSION,
  "version" : VERSION,
  "expires" : EXPIRES,
  "targets" : TARGETS,
  ("delegations" : DELEGATIONS)
}
```

where
* `SPEC_VERSION`, `VERSION` and `EXPIRES` are the same as is described for the root.json file.
* `TARGETS` is an object whose format is the following:
```
{
  TARGETTAG : {
      "length" : LENGTH,
      "hashes" : HASHES,
      ("annotation" : ANNOTATION) }
  , ...
}
```
* `TARGETTAG` is a string giving the fully qualified reference for an image.
It is allowed to have a `TARGETS` object with no `TARGETTAG`
elements if no target files are available.
* `LENGTH` is an integer length in bytes of the target file at TARGETTAG.
* `HASHES` is a dictionary that specifies one or more hashes of the target file at TARGETTAG, with a string describing the cryptographic hash function as key and HASH as defined for METAPATHS. For example: { "sha256": HASH, ... }.
* `ANNOTATION` is an object. If defined, the elements and values of the ANNOTATION object will be made available to the client application. The format of the ANNOTATION object is opaque to the framework, which only needs to know that the "annotation" attribute maps to an object. The ANNOTATION object may include version numbers, dependencies, requirements, or any other data that the application wants to include to describe the file at TARGETTAG. The application may use this information to guide download decisions.
* `DELEGATIONS` is an object whose format is the following:
```
{
  "keys" : {
      KEYID : KEY,
      ...
  },
  "roles" : [
    {
      "name": ROLENAME,
      "keyids" : [ KEYID, ... ] ,
      "threshold" : THRESHOLD,
      ("path_hash_prefixes" : [ HEX_DIGEST, ... ] |
      "succinct_hash_delegations" : {
         "delegation_hash_prefix_len" : BIT_LENGTH,
         "bin_name_prefix" : NAME_PREFIX
         } |
      "paths" : [ PATHPATTERN, ... ]),
      "terminating": TERMINATING,
    },
    ...
  ]
}
```
* `KEYID` and `KEY` are the same as is described for the root.json file.
* `ROLENAME` is a string giving the name of the delegated role. For example, "projects".
* `TERMINATING` is a boolean indicating whether subsequent delegations should be considered.

As explained in the [Diplomat paper
](https://theupdateframework.io/papers/protect-community-repositories-nsdi2016.pdf),
terminating delegations instruct the client not to consider future trust
statements that match the delegation's pattern, which stops the delegation
processing once this delegation (and its descendants) have been processed.
A terminating delegation for a package causes any further statements about a
package that are not made by the delegated party or its descendants to be
ignored.

* In order to discuss target paths, a role MUST specify only one of the "path_hash_prefixes", "succinct_hash_delegations", or "paths" attributes, each of which we discuss next.
   * `path_hash_prefixes` :: A list of HEX_DIGESTs used to succinctly describe a set of target paths. Specifically, each HEX_DIGEST in "path_hash_prefixes" describes a set of target paths; therefore, "path_hash_prefixes" is the union over each prefix of its set of target paths. The target paths must meet this condition: each target path, when hashed with the SHA-256 hash function to produce a 64-byte hexadecimal digest (HEX_DIGEST), must share the same prefix as one of the prefixes in "path_hash_prefixes". This is useful to split a large number of targets into separate bins identified by consistent hashing.
   * `succinct_hash_delegations` :: A dictionary containing succinct descriptions of delegations to a set of bins, similar to those used in `path_hash_prefixes`. This field represents delegations to 2^BIT_LENGTH bins that use the keyid, threshold, path, and termination status from this delegation. The path_hash_prefixes and name for each bin will be determined using the BIT_LENGTH and NAME_PREFIX. BIT_LENGTH must be an integer between 1 and 32 (inclusive). The rolename for each bin will be structured as NAME_PREFIX-COUNT where COUNT is a hexadecimal value between 0 and 2^BIT_LENGTH-1 (inclusive) that represents the bin number. If the succinct_hash_delegations field is present in a delegation, the name field will not be used as a rolename, and so is not required.
   * `paths` :: A list of strings, where each string describes a path that the role is trusted to provide. Clients MUST check that a target is in one of the trusted paths of all roles in a delegation chain, not just in a trusted path of the role that describes the target file. PATHPATTERN can include shell-style wildcards and supports the Unix filename pattern matching convention. Its format may either indicate a path to a single file, or to multiple paths with the use of shell-style wildcards. For example, the path pattern "targets/*.tgz" would match file paths "targets/foo.tgz" and "targets/bar.tgz", but not "targets/foo.txt". Likewise, path pattern "foo-version-?.tgz" matches "foo-version-2.tgz" and "foo-version-a.tgz", but not "foo-version-alpha.tgz". To avoid surprising behavior when matching targets with PATHPATTERN, it is RECOMMENDED that PATHPATTERN uses the forward slash (/) as directory separator and does not start with a directory separator, akin to TARGETPATH.

Prioritized delegations allow clients to resolve conflicts between delegated roles that share responsibility for overlapping target paths by considering metadata in order of appearance of delegations; we treat the order of delegations such that the first delegation is trusted over the second, the second over the third, and so on. Likewise, the metadata of the first delegation will override that of the second delegation, the metadata of the second delegation will override that of the third one, etc. In order to accommodate prioritized delegations, the "roles" key in the DELEGATIONS object above points to an array of delegated roles, rather than to a hash table.

The metadata files for delegated target roles have the same format as the top-level targets.json metadata file.

### Timestamp

Timestamp metadata will contain the following in the `ROLE` field:
```
{
  "_type" : "timestamp",
  "spec_version" : SPEC_VERSION,
  "version" : VERSION,
  "expires" : EXPIRES,
  "meta" : METAFILES
}
```
where
* `SPEC_VERSION`, `VERSION` and `EXPIRES` are the same as is described for the root.json file.
* `METAFILES` is the same as described for the snapshot.json file. In the case of the timestamp.json file, this MUST only include a description of the snapshot.json file.
