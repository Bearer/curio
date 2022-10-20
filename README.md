<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="./docs/assets/img/curio-logo-dark.svg">
  <img alt="Curio" src="./docs/assets/img/curio-logo-light.svg">
</picture>

<br />

[![GitHub Release][release-img]][release]
[![Test][test-img]][test]
[![GitHub All Releases][github-all-releases-img]][release]

</div>

# Curio

Curio is a static code analysis tool (SAST) dedicated to data security. It scans your source code to discover sensitive data flows (PHI, PII, PD as well as Data stores, internal and external APIs) and data security risks (leaks, missing encryption, third-party sharing, etc).
You can use Curio as a CLI or add it to your CI/CD pipeline.

Curio currently supports Java and Ruby stacks, more will be added in the coming months (cf roadmap).

Curio also supports the following structured file definition:

- SQL
- Open API
- Protobuf
- GraphQL

Curio detections, classifications and policies are fully customizable.

Curio also powers [Bearer](https://www.bearer.com), the developer-first platform to proactively find and fix data security risks and vulnerabilities across your development lifecycle

[Read the docs](/docs/) to learn more and get started.

## Getting started

Scan your first project in X minutes or less.

## Installation

Curio is available as a standalone executable binary. The latest release is available on the [releases tab](https://github.com/Bearer/curio/releases/latest), or use one of the methods below.

### Install Script

:warning: **Not working till public** :warning:

This script downloads the Curio binary automatically based on your OS and architecture.

```bash
curl -sfL https://raw.githubusercontent.com/Bearer/curio/main/contrib/install.sh | sh
```

_Defaults to `./bin` as a bin directory and to the latest releases_

If you need to customize the options, use the following to pass parameters:

```bash
curl -sfL https://raw.githubusercontent.com/Bearer/curio/main/contrib/install.sh | sh -s -- -b /usr/local/bin
```

### Binary

Download the archive file for your operating system/architecture from [here](https://github.com/Bearer/curio/releases/latest/). Unpack the archive, and put the binary somewhere in your $PATH (on UNIX-y systems, /usr/local/bin or the like). Make sure it has execution bits turned on:

```bash
chmod +x ./curio
```

### Scan your project

Run `curio scan` on a project directory:

```bash
curio scan /path/to/your_project
```

or a single a file:

```bash
curio scan ./curio-ci-test/Pipfile.lock
```

<!-- TODO: insert sample output or video here -->

Additional options for using and configuring the scan command can be found in the [scan documentation](docs/reference/commands.md#scan);

## Development

### Installation

Install modules:

```bash
go mod download
```

### Testing

Running classification tests:

```bash
go test ./pkg/classification/... -count=1
```

Running a single specific test:

```bash
go test -run ^TestSchema$ ./pkg/classification/schema -count=1
```

---

[test]: https://github.com/Bearer/curio/actions/workflows/test.yml
[test-img]: https://github.com/Bearer/curio/actions/workflows/test.yml/badge.svg
[release]: https://github.com/Bearer/curio/releases
[release-img]: https://img.shields.io/github/release/Bearer/curio.svg?logo=github
[github-all-releases-img]: https://img.shields.io/github/downloads/Bearer/curio/total?logo=github
