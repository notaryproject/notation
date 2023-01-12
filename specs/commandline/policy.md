# notation policy

## Description

Use `notation policy` command to manage trust policies.

Trust policies are configured in a `JSON` document named `trustpolicy.json`. The goal of `notation policy` command is that users don't need to create or update `trustpolicy.json` manually.

TODO: explain the verification level

## Outline

### notation policy command

```text
Manage trust policies for signature verification.

Usage:
  notation policy [command]

Available Commands:
  add           Add trust policies.
  delete        Delete trust policies.
  init          Initialize trust policies.
  list          List trust policies.
  show          Show details of a trust policy.
  update        Update trust policies

Flags:
  -h, --help   help for policy
```

### notation policy add

```text
Add trust policies.

Example - Add a trust policy that trusts any identities under specified trust store to validate any artifacts:
  notation add --name wabbit-network-build --ts ca:wabbit-network-build

Example - Add a trust policy that trusts any identities under specified trust store to validate artifacts stored in specified repository:
  notation add --name wabbit-network-build --scope localhost:5000/build/net-monitor --ts ca:wabbit-network-build

Example - Add a trust policy that trusts specified identity (certificate file) to validate artifacts stored in specified repository:
  notation add --name wabbit-network-build --scope localhost:5000/build/net-monitor --ts ca:wabbit-network-build --id-cert wabbit-network-build.crt

Example - Add a trust policy that trusts specified identity (x509 subject) to validate artifacts stored in specified repository:
  notation add --name wabbit-network-build --scope localhost:5000/build/net-monitor --ts ca:wabbit-network-build --id "x509.subject: CN=SecureBuilder"

Example - Add trust policies from a json file, which are merged with existing trust policies:
  notation add --input @wabbit-network.json

Example - Add trust policies from an JSON object, which are merged with existing trust policies:
  notation update --input {"version": "1.0", "trustPolicies": [{"name": "wabbit-networks-images",...}]}

Usage:
  notation policy add [flags]

Flags:
  -h, --help     help for add
      --input    input as a json file or a json object
```

### notation policy delete

```text
Delete trust policies. User cannot delete all the trust policies, at least one trust policy should be configured for signature verification.

Example - Delete one trust policy by name
  notation policy delete wabbit-network-build

Example - Delete multiple trust polices by names
  notation policy delete wabbit-network-build wabbit-network-publish

Usage:
  notation policy delete [flags] <name>...

Flags:
  -h, --help   help for delete
```

### notation policy init

```text
Initialize trust policies.

Example - Init the trust policy that trusts any identities under certain trust store to verify artifacts stored in specified repositories:
  notation policy init --repo localhost:5000/build/net-monitor --ts ca:wabbit-network-build

Example - Init the trust policy from a json file, which overrides all existing trust policies if confirmed:
  notation policy init --input @wabbit-network.json

Example - Init the trust policy from a json object, which overrides all existing trust policies if confirmed:
  notation policy init --input {"version": "1.0", "trustPolicies": [{"name": "wabbit-networks-images",...}]}

Usage:
  notation policy init [flags]

Flags:

  -h, --help     help for list
      --input    input as a json file or a json object
      --repo     repository that trust policy is applicable for
      --ts       trust store that contains the trust identities
```

### notation policy list

```text
List trust policies.

Example - List all the trust policies:
  notation policy list

Example - List trust policies for the specified artifact, e.g. container image:
  notation policy list --ref localhost:5000/net-monitor@sha256:xxx

Example - List trust policies with specified trust store configured:
  notation policy list --ts ca:wabbit-network

Usage:
  notation policy list [flags]

Aliases:
  list, ls

Flags:
      --ref      list the trust policies for the artifacts
  -h, --help     help for list
      --ts       list the trust policies with specified trust stores
```

### notation policy show

```text
Show details of trust policies.

Example - Show details of the trust policy by name
  notation policy show wabbit-network-build

Usage:
  notation policy show [flags] <name>

Flags:
  -h, --help   help for show
```

### notation policy update

```text
Update the existing trust policies.

Example - Update trust policy from an input file, which overrides all the existing policies:
  notation update --input @wabbit-network.json

Example - Update trust policy from an input JSON object, which overrides all the existing policies:
  notation update --input {"version": "1.0", "trustPolicies": [{"name": "wabbit-networks-images",...}]}

Example - Update the registry scopes for a trust policy:
  notation update --name wabbit-network-build --scope localhost:5000/build/net-monitor --scope localhost:5000/build/nginx

Example - update the verification level for a trust policy:
  notation update --name wabbit-network-build --level-audit

Example - Update the trust stores for a trust policy:
  notation update --name wabbit-network-build --ts ca:wabbit-network-build --ts ca:wabbit-network-publish

Example - Update the trust identities for a trust policy by setting specified x509 subjects:
  notation update --name wabbit-network-build --id "x509.subject: C=US, ST=WA, L=Seattle, O=wabbit-networks.io, OU=Finance, CN=SecureBuilder" --id "x509.subject: C=US, ST=WA, L=Seattle, O=wabbit-networks.io, OU=Marketing, CN=SecureBuilder"

Example - Update the trust identities for a trust policy by setting specified certificate name in the trust stores:
  notation update --name wabbit-network-build --id-cert wabbit-network-build.crt

Usage:
  notation policy update [flags]

Flags:
  -h, --help   help for update
```

## Usage
