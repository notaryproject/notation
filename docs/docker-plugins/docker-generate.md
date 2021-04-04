# docker-generate
Docker CLI plugin to generate metadata **offline**.

## Build and Install
This plugin requires [golang](https://golang.org/dl/) with version `>= 1.16`.

To build and install, run
```
go build -o ~/.docker/cli-plugins/docker-generate ./cmd/docker-generate
```
Finally, follow the [documentation](https://docs.docker.com/engine/reference/commandline/cli/#experimental-features) to enable experimental features for Docker CLI.

## Instructions
### Generate manifest
To generate a manifest of an image referenced by `<reference>`, run
```
docker generate manifest <reference>
```
For example,
```
docker build -t myapp:1.0 .
docker generate manifest myapp:1.0
```

#### Output to files
If a file is desired instead of standard output, try
```
docker generate manifest <reference> > manifest.json
```
or
```
docker generate manifest -o manifest.json <reference>
```

#### From `docker save`
Manifest is also possible to be generated from a tar file saved via `docker save`.
```
docker save <reference> > save.tar
cat save.tar | docker generate manifest
```
Basically,
```
docker save <reference> | docker generate manifest
```
is equivalent to 
```
docker generate manifest <reference>
```

## Known Limitation
The current version of this plugin can only accept one `reference` at a time.
