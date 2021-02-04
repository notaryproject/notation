The TUF + Notary design will require a few additions to the OCI specification.

The specification will need to support the following media types:

* A metadata type that refers to images or other metadata files. This can be a
general type for any metadata, or could be broken into the following subtypes,
described in more detail in the design:
  * Targets metadata
  * Snapshot metadata
  * Timestamp metadata
  * Root metadata

  These metadata files will need to be accessed by name so that the user can
  download the latest versions of them. This access by name could be achieved
  using existing tag resolution mechanisms.
