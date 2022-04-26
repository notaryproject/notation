# Contributing

Notary v2 is [Apache 2.0 licensed](https://github.com/notaryproject/nv2/blob/main/LICENSE) and
accepts contributions via GitHub pull requests. This document outlines
some of the conventions on to make it easier to get your contribution
accepted.

We gratefully welcome improvements to issues and documentation as well as to
code.

## Certificate of Origin

By contributing to this project, you agree to the Developer Certificate of
Origin (DCO). This document was created by the Linux Kernel community and is a
simple statement that you, as a contributor, have the legal right to make the
contribution.

We require all commits to be signed. By signing off with your signature, you
certify that you wrote the patch or otherwise have the right to contribute the
material by the rules of the [DCO](https://github.com/apps/dco):

`Signed-off-by: Jane Doe <jane.doe@example.com>`

The signature must contain your real name *(sorry, no pseudonyms or anonymous contributions)*.
If your `user.name` and `user.email` are configured in your Git config,
you can sign your commit automatically with `git commit -s`.

As our project is about code-signing we also highly appreciate it if you sign your commits using a GPG key. :smile:

## Communications

For realtime communications we use Slack: To join the conversation, simply
join the [CNCF](https://slack.cncf.io/) Slack workspace and use the
[#notary-v2](https://cloud-native.slack.com/messages/notary-v2/) channel.

To discuss ideas and specifications we use [Github
Discussions](https://github.com/notaryproject/notaryproject/discussions).

## Understanding Notary v2

This project is composed of:

- [notation](https://github.com/notaryproject/notation): The Notary v2 CLI and Docker plugins
- [notation-go-lib](https://github.com/notaryproject/notation-go): A collection of libraries for supporting Notation sign, verify of oci artifacts. Based on Notary V2 standard.
- [notaryproject](https://github.com/notaryproject/notaryproject): The Notary v2 requirements and scenarios to frame the scope of the Notary project
- [tuf-notary](https://github.com/notaryproject/tuf): Integration of Notary v2 and TUF

Also consider checking out our [roadmap](https://github.com/notaryproject/roadmap).

### Understanding the code

We are using the following [project-layout](https://github.com/golang-standards/project-layout).

### How to run the test suite

Prerequisites:

- go >= 1.17

You can run the unit tests by simply doing

```bash
make test
```

## Acceptance policy

These things will make a PR more likely to be accepted:

- a well-described requirement
- tests for new code
- tests for old code!
- new code and tests follow the conventions in old code and tests
- a good commit message ([see below](#format-of-the-commit-message))
- all code must abide [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- names should abide [What's in a name](https://talks.golang.org/2014/names.slide#1)
- code must build on both Linux, Windows and Darwin, via plain `go build`
- code should have appropriate test coverage and tests should be written
  to work with `go test`

In general, we will merge a PR once one maintainer has endorsed it.
For substantial changes, more people may become involved, and you might
get asked to resubmit the PR or divide the changes into more than one PR.

### Format of the Commit Message

We prefer the following rules for good commit messages:

- Limit the subject to 50 characters and write as the continuation
  of the sentence "If applied, this commit will ..."
- Explain what and why in the body, if more than a trivial change;
  wrap it at 72 characters.

The [following article](https://chris.beams.io/posts/git-commit/#seven-rules)
has some more helpful advice on documenting your work.
