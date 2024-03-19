# gh-action-cache-nuker

As the name implies this script is used for deleting github action caches.

## Installation

To install the program [Go](https://go.dev) is required.

After installing Go, the program can be installed by running:

```sh
go install github.com/hougesen/gh-action-cache-nuker@latest
```

The program can now be run by either calling the program directly (Most likely `$HOME/go/bin/gh-action-cache-nuker`) or by setting up a path to the go install path in your `.bashrc` and calling the name of the program (`gh-action-cache-nuker`).

```sh
# .bashrc
export PATH=${PATH}:$(go env GOPATH)/bin
```

## Usage

A [GitHub token](https://github.com/settings/tokens) is needed to run the script.

The token must have the following repository permissions:

- Read access to administration and metadata
- Read and Write access to actions

It should of course also have access to the repository you wish to run the script against.

### Repository

```sh
gh-action-cache-nuker repo <NAME_OF_USER_OR_ORGANIZATION/NAME_OF_REPOSITORY> <GITHUB_ACTION_TOKEN>
```

### User/Organization

```sh
gh-action-cache-nuker org <NAME_OF_USER_OR_ORGANIZATION> <GITHUB_ACTION_TOKEN>
```
