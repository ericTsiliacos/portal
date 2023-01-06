# 🌀 Portal

https://erictsiliacos.medium.com/remote-pairing-with-portal-ba107fcb824f

### Installation
```brew install erictsiliacos/tap/portal```

# Usage

> You can use the `--help` option to get more details about the commands and their options

## Getting Started

Set up the same pairing using [git-duet](https://github.com/git-duet/git-duet) or [git-together](https://github.com/kejadlen/git-together)

## Push

> Push local changes to a remote branch based on your pairing

```bash
portal push
```

Options

```
 -h, --help         displays usage information of the application or a command (default: false)
 -s, --strategy     git-duet, git-together (default: auto)
 -v, --verbose      verbose output (default: false)
```

## Pull

> Pull changes from the portal branch name and clean up the temporary branch

```bash
portal pull
```

Options

```
 -h, --help         displays usage information of the application or a command (default: false)
 -s, --strategy     git-duet, git-together (default: auto)
 -v, --verbose      verbose output (default: false)
```

### Environment Variables

Setting `PORTAL_COMMIT_MESSAGE` to a string of your choice will add to the commit message that portal creates
if there exists git commit message hooks that need satisfying

e.g. ```export PORTAL_COMMIT_MESSAGE="<message goes here>"```

### Logs

`~/.portal/Logs/info.log`
  
### Supports
- [git-duet](https://github.com/git-duet/git-duet)
- [git-together](https://github.com/kejadlen/git-together)

## Contribute

### Bats
Bats is a bash testing framework, used here for integration tests. This can be installed with homebrew.

```brew bundle```

### Testing
```go test ./...```

```bats ./bats```

## Credit 

Idea: George Dean ([@gdean123](https://github.com/gdean123))
