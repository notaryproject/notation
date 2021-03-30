# docker-nv2
Docker CLI plugin to demonstrate Notary v2 integration with Docker CLI, leveraging the [notary v2 library](https://github.com/notaryproject/notary).

## Build and Install
This plugin requires [golang](https://golang.org/dl/) with version `>= 1.14`.

To build and install, run
```
go build -o ~/.docker/cli-plugins/docker-nv2 ./cmd/docker-nv2
```

## Instructions
### Create Alias
For better demonstration experience, it is suggested to create the following alias in your shell:
```bash
alias docker="docker nv2"
```
or if you are using PowerShell on Windows,
```powershell
function docker { cmd /c docker nv2 $args }
```

### Example Run
On the producer machine:
```bash
docker notary --enabled
docker build -t $image .
docker notary sign --key identity.pem --cert identity.crt $image
docker push $image
```

On the consumer machine:
```bash
docker notary --enabled
docker pull $image
```
It may fail since the producer machine may use a self-signed certificate, or invalid certificates detected.
See [configurations](#configurations) for more details.

## Configurations
The config file for notary is default at `~/.docker/nv2.json`.
The intermediate signatures are stored at `~/.docker/nv2/`.

The config file looks like
```json
{
    "enabled": true,
    "verificationCerts": [
        "path/to/the/certs/for/verification"
    ]
}
```
To pull images properly, certification paths are required to be provided at `verificationCerts`.
It is suggested to use absolute paths.
