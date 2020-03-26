# Cuckoo

Cuckoo is a command line tool to simplify CI pipelines. It is currently focused on GitLab CI and deployment to Kubernetes clusters provisioned via the Google Cloud Platform.

Cuckoo enables moving from error-prone bash scripts to a powerful Go CLI tool that minimizes the time spent on debugging CI pipelines. Additionally, it has a minimal footprint (the image has a size below 60 MB at the moment) and thus speeds up CI pipelines and saves bandwidth.

## Features

Cuckoo provides the following set of commands:

* `auth`: Checks for authentication against multiple components and performs a login from credentials given by environment variables if required (e.g. SSH daemon, Docker registry, Google Cloud Platform).
* `build`: Builds a Docker container and optionally pushes it to a registry (with multiple tags). Builds can be performed using a (remote) BuildKit daemon.
* `decrypt`: Automatically decrypt all files matching some pattern using Mozilla's [Sops](https://github.com/mozilla/sops).
* `deploy`: Deploy a Helm chart or single Kubernetes manifests to a Kubernetes cluster.
* `provision`: Provision infrastructure using Terraform.
* `publish`: Upload static files to an object storage bucket to be served as static website.

More details explanations for the commands can be retrieved by installing the `cuckoo` command and running `cuckoo help <command>`.

## Installation

Currently, Cuckoo is available via Docker and Homebrew, i.e. on MacOS. In any other case, it must be built from source.

### Docker

Cuckoo is hosted on Docker Hub and can therefore be simply referenced under the following name:

```
borchero/cuckoo
```

### MacOS

Cuckoo resides in a [custom Homebrew tap](https://github.com/borchero/homebrew) and must therefore be installed via the following two commands:

```bash
brew tap borchero/tools
brew install cuckoo
```

*You do not need to have Go installed as a static binary will be shipped.*

### Build from Source

Building from source is as simple as cloning the repository and running the following command inside the `source` directory:

```bash
go install
```

*Make sure you have Go version 1.14 installed.*
