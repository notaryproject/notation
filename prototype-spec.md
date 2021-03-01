# Notary v2 Specification

This specification describes the file formats, queries, and processes required to use TUF in registries. Primarily those requirements are that metadata for each of the TUF roles root, snapshot, timestamp, and targets be uploaded to a registry and that this metadata can be queried by users. Details about each of these roles can be found in the [design document](https://github.com/notaryproject/nv2/blob/prototype-tuf/tuf-design.md) and the [TUF specification](https://github.com/theupdateframework/specification/blob/master/tuf-spec.md#the-update-framework-specification).

## Root metadata

Root metadata should be signed using an offline process. It can then be uploaded to the registry by running `nv2 upload root`.

## Signing a manifest

To upload a manifest signed with Notary v2, a developer must:
* Generate and sign targets metadata by running `nv2 sign imagename keyloc --role targetsrolename` where `imagename` is the manifest to be signed, `targetsrolename` is the name of the top-level or delegated targets role the developer is a part of, and `keyloc` is the location of the developer's private key.
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

Snapshot metadata may be stored in a [snapshot Merkle tree](https://github.com/theupdateframework/taps/blob/master/tap16.md) instead of a single metadata file to reduce the size of downloaded metadata. To do so, implementations should follow the same process as described here, but upload all snapshot Merkle metadata files and add the Merkle root hash to timestamp metadata. Implementations that use snapshot Merkle trees should support auditors as described in the TAP.

## User download

When installing an image from a registry, a user may run `nv2 download imagename`. This command must download the following from the registry:
* Root metadata. The client should download the most recent root metadata, as well as all previous versions until they reach their currently trusted root version.
* The most recent timestamp metadata.
* The most recent snapshot metadata.
* the most recent top-level targets metadata.
* Relevant delegated targets metadata. If the top-level targets metadata delegated the queried namespace to another targets metadata file, that file should be downloaded as well. Be aware that there may be multiple levels of delegated targets metadata.
* The requested image.

Metadata files may be stored on a different repository than the downloaded image. If so, the repository should contain a placeholder that redirects the user to download the metadata from the appropriate location. For example, if root metadata is shared between multiple repositories, it may be stored in a central location, with placeholders in each repository pointing the the location of the root metadata.

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

## Example Metadata

### Root

```
{
 "signatures": [
  {
   "keyid": "638a9a238471b2b6da3ee6b5d8cd515f59f07de8f6e50b2098935e0d00efc360",
   "sig": "cd06a26020bcdd932d671d6660f7a0474314834d9017b3925eafe1da1b5d0531d12d6c2885c4ea8d9c3e572d726ef9760361e1e012b0ed1feaa0d9187c094ea362124f3fe6d13ced12a7201a4b99ad56cff7b8834bc2fac72135c281a44467deaeb72ff44b94619f59a5387dee9a39178b715a4e4159f66e019f366ddd504f1af6ef7bc25bb35afed7552fe8806a35b2787bf1be1f4578e546cf8209362e1cb1997daad82b26a332e2a4087ef2ab4f9d83d451c0f0316c0862eb5abc9e71d898496de96de820ba6f74b84244d2c8a692da8a47f665a110345ea2b3d95abf6ff0ec226652f0a56aa3a9cedfe2baf0b0d1a46106dcaaabf8602ceb7ba2514113ef"
  },
  {
   "keyid": "824b2ea397699d84e85d1caf21e14046606aed6e24a70e559c486b94c261da29",
   "sig": "1d6b16241be28f4a1883d0dd8e707fac29d4f6aa1b7e62f3d37e3de93ae21feaddc36593d45e14f8ac23fdfe4ae543a0f7e2ba34009579dcdfadad398de818d466453c98dbc6fc9cc69ec74c42dd37a76be5281b943e21375689cc2a4860dd7b2eb24842997b1637706f072416647906f2a852fac66fca3ba2a0a2d8f9a2912430e2cce6d8f1fc0b48cb4e6bc3433aa567eba6f30bb1e73819052f9670a118fdf8e025d0357226f7fae824a1e26bb4e90e09572e7ba0b691b70ccd6bde239e0bb7b4ffeb869ed91c568d3566991bf4ac75854e96507a1cfce3ee4f3d18df5435c4a2c599788085fb56e59c37838a6eb061e8410fc5e60122566d35f7229e0c7e"
  }
 ],
 "signed": {
  "_type": "root",
  "consistent_snapshot": false,
  "expires": "2021-11-12T00:23:43Z",
  "keys": {
   "156b909894a0cee7ede09baeab7e42e0b69507e7b849843164379893e84c6245": {
    "keyid_hash_algorithms": [
     "sha256",
     "sha512"
    ],
    "keytype": "ed25519",
    "keyval": {
     "public": "f0648470c30395377550af67d173c494721af5b648a03500ccee231ee77104fa"
    },
    "scheme": "ed25519"
   },
   "4ef5bb3d874885be40518d233c762fbc455e382989b4435f74379f70212ffe1a": {
    "keyid_hash_algorithms": [
     "sha256",
     "sha512"
    ],
    "keytype": "ed25519",
    "keyval": {
     "public": "2ec03df3e4ffb9058d3bfca640a9e5d4cebe4d8159168373598541252dbcbcd8"
    },
    "scheme": "ed25519"
   },
   "638a9a238471b2b6da3ee6b5d8cd515f59f07de8f6e50b2098935e0d00efc360": {
    "keyid_hash_algorithms": [
     "sha256",
     "sha512"
    ],
    "keytype": "rsa",
    "keyval": {
     "public": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzdBB9mDy+fBtdst5Izsh\nJaYkWe1njs5I9rwlDhYGy0wDjAwaI+YQAssAiLEdVYavSY4s02Y+EWBdFRJKxWdM\nAmtTMCpD8ZjkJkIaedJljpg0Qh2hdQYhMERKjOMKhZLK0EKHUtziXKT1vNXv1J95\nWQjM5c+eCOyzoxPwym3Bo4vl/+DsiwZixWtKrk3imFvnEJCrJLGZecUG2sZOs+W3\nQP5vYnZc+d+XhBmNGXWJrO5h52lRF9BVdXQhQxMKQhpbUkuceq9Svu0AQ+1Jdtro\nJQ4UVNKBXjCVzxJ4nD7O7oNMNUpe1K+oaHGZi5Jo5q0qVVm17yb3k71Tq5dhHB3+\n3QIDAQAB\n-----END PUBLIC KEY-----"
    },
    "scheme": "rsassa-pss-sha256"
   },
   "824b2ea397699d84e85d1caf21e14046606aed6e24a70e559c486b94c261da29": {
    "keyid_hash_algorithms": [
     "sha256",
     "sha512"
    ],
    "keytype": "rsa",
    "keyval": {
     "public": "-----BEGIN PUBLIC KEY-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtWQGNB6gTGk4kguemL/M\n6bgw4u4GMsgxNXXLkksuE5D4o2UbHGknUTdQVsZZS1Yq6k5sszawoMkbxL4uOmDV\nxM73P7TlwII3xTTw/bijNnHYnQ72lOceRikgSWM6gd4puCnKQe/tlMp54pd27vZc\nLKgALlznaJiFJL7BMttk0TWDX1IPo48beE5lSXW6u1n5ZQyEtv4lr0M7c2AyWWdk\n4LhZYHSGBWDyj3fHcfMc6nqHDmMObYMC+GX9YO1NPpjxOt2j485DeuRNd+OEuY/X\nPr3Ri5QaIZtmmD4/u8HuGcfcMOesjFPxHkLtpvf5wBUY1A6NmDO7OZJwDnOaQh3k\njwIDAQAB\n-----END PUBLIC KEY-----"
    },
    "scheme": "rsassa-pss-sha256"
   },
   "a05eb982a411de397adc3e553cbac8c057346750477ccfca02890ede8438e3a8": {
    "keyid_hash_algorithms": [
     "sha256",
     "sha512"
    ],
    "keytype": "ed25519",
    "keyval": {
     "public": "365051c1b9ce6621cdb0f64af2d938e0bc2e2e9ddf9665e276ae8fef985e1fa5"
    },
    "scheme": "ed25519"
   }
  },
  "roles": {
   "root": {
    "keyids": [
     "638a9a238471b2b6da3ee6b5d8cd515f59f07de8f6e50b2098935e0d00efc360",
     "824b2ea397699d84e85d1caf21e14046606aed6e24a70e559c486b94c261da29"
    ],
    "threshold": 2
   },
   "snapshot": {
    "keyids": [
     "4ef5bb3d874885be40518d233c762fbc455e382989b4435f74379f70212ffe1a"
    ],
    "threshold": 1
   },
   "targets": {
    "keyids": [
     "156b909894a0cee7ede09baeab7e42e0b69507e7b849843164379893e84c6245"
    ],
    "threshold": 1
   },
   "timestamp": {
    "keyids": [
     "a05eb982a411de397adc3e553cbac8c057346750477ccfca02890ede8438e3a8"
    ],
    "threshold": 1
   }
  },
  "spec_version": "1.0.0",
  "version": 1
 }
}
```

### Timestamp

```
{
 "signatures": [
  {
   "keyid": "a05eb982a411de397adc3e553cbac8c057346750477ccfca02890ede8438e3a8",
   "sig": "57078fda6b46248ed4a7195972ac6a26583357c8e563879ac6c6057bd0f07c628e00979d62ff7236fcd56f9fed321ab0e2d132766871f2bbd13b86b3032eec0e"
  }
 ],
 "signed": {
  "_type": "timestamp",
  "expires": "2080-10-28T12:08:00Z",
  "merkle_root": "86411150f9003dc98ee86bcee2b9eeb57bb5ffd9cc822e38255f4ef091036153",
  "meta": {
   "snapshot.json": {
    "hashes": {
     "sha256": "71e8c5a5ab141c698d26eb14f79fbce8cd8313fa3859771dc48eb56134187529",
     "sha512": "c09a776e0ab9a3d9838df90a63cf0ff5a86710cd6050358259b66f75fe3f232625cb2e15090ad8812f698c87960546136deda8f8e17e691099df6ef283ae947e"
    },
    "length": 1251,
    "version": 26
   }
  },
  "spec_version": "1.0.0",
  "version": 26
 }
}
```

### Snapshot

```
{
 "signatures": [
  {
   "keyid": "4ef5bb3d874885be40518d233c762fbc455e382989b4435f74379f70212ffe1a",
   "sig": "82dcb500ea3edd47a33820e4c5902e7dae725fafbb423f6ca6074a97862bcaa8eec637b1c65bd6c7159c8f108f0726b45d33b1d0d39b2fc437b7dcdd55b2f507"
  }
 ],
 "signed": {
  "_type": "snapshot",
  "expires": "2080-10-28T12:08:00Z",
  "meta": {
   "my_repo.json": {
    "version": 4
   },
   "repository0.json": {
    "version": 1
   },
   "repository1.json": {
    "version": 1
   },
   "repository10.json": {
    "version": 1
   },
   "repository11.json": {
    "version": 1
   },
   "repository12.json": {
    "version": 1
   },
   "repository13.json": {
    "version": 1
   },
   "repository14.json": {
    "version": 1
   },
   "repository15.json": {
    "version": 1
   },
   "repository2.json": {
    "version": 1
   },
   "repository3.json": {
    "version": 1
   },
   "repository4.json": {
    "version": 1
   },
   "repository5.json": {
    "version": 1
   },
   "repository6.json": {
    "version": 1
   },
   "repository7.json": {
    "version": 1
   },
   "repository8.json": {
    "version": 1
   },
   "repository9.json": {
    "version": 1
   },
   "targets.json": {
    "version": 11
   }
  },
  "spec_version": "1.0.0",
  "version": 26
 }
}
```

### Targets

```
{
 "signatures": [],
 "signed": {
  "_type": "targets",
  "delegations": {
   "keys": {
    "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f": {
     "keyid_hash_algorithms": [
      "sha256",
      "sha512"
     ],
     "keytype": "ed25519",
     "keyval": {
      "public": "76880651c7d186ffb9370d9905f4bd5fc33d7fa28cb5885d6dd36a75124eab7d"
     },
     "scheme": "ed25519"
    }
   },
   "roles": [
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "my_repo",
     "paths": [
      "my_repo/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository0",
     "paths": [
      "repository0/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository1",
     "paths": [
      "repository1/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository2",
     "paths": [
      "repository2/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository3",
     "paths": [
      "repository3/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository4",
     "paths": [
      "repository4/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository5",
     "paths": [
      "repository5/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository6",
     "paths": [
      "repository6/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository7",
     "paths": [
      "repository7/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository8",
     "paths": [
      "repository8/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository9",
     "paths": [
      "repository9/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository10",
     "paths": [
      "repository10/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository11",
     "paths": [
      "repository11/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository12",
     "paths": [
      "repository12/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository13",
     "paths": [
      "repository13/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository14",
     "paths": [
      "repository14/*"
     ],
     "terminating": false,
     "threshold": 1
    },
    {
     "keyids": [
      "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f"
     ],
     "name": "repository15",
     "paths": [
      "repository15/*"
     ],
     "terminating": false,
     "threshold": 1
    }
   ]
  },
  "expires": "2021-02-11T02:02:33Z",
  "spec_version": "1.0.0",
  "targets": {},
  "version": 11
 }
}
```

### Delegated Targets

```
{
 "signatures": [
  {
   "keyid": "09d2912b04d0965464872958571927e7cdf67daf86db30a75205cd64c154c91f",
   "sig": "16237d40ddfc0957b07667b29170b634cb9206a55deb7eb3630a840f97484eced9659c962fa11215f49aa22c04b12279a738ca637220b8a7f3ffa071cc6e140b"
  }
 ],
 "signed": {
  "_type": "targets",
  "delegations": {
   "keys": {},
   "roles": []
  },
  "expires": "2021-02-11T02:02:33Z",
  "spec_version": "1.0.0",
  "targets": {
   "my_repo/image1": {
    "custom": {},
    "hashes": {
     "sha256": "3c6232e288788106ca68e47f84263fbafe14d418cee6528b52446c3784c77e27",
     "sha512": "5521c3af40060a53bd04f647e10a8d02b136503a724046c5fca24cfcbf34ddf2d56270732c6e8c8db634e6bfea534872e1ae7c9d415674485c3d0b79c7f721b5"
    },
    "length": 27
   }
  },
  "version": 4
 }
}
```
