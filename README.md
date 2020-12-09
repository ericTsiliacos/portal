# 🌀 Portal

# Usage

> You can use the `--help` option to get more details about the commands and their options

## Push

Pushes local changes to a remote branch based on your pairing

```bash
portal push
```

Options

```
 -h, --help         displays usage information of the application or a command (default: false)
 -p, --patch        create and stash patch (default: false)
 -s, --strategy     git-duet, git-together (default: auto)
 -v, --verbose      verbose output (default: false)
```

## Pull

Pulls the changes at the pair branch name and cleans up the temporary branch

```bash
portal pull
```

Options

```
 -h, --help         displays usage information of the application or a command (default: false)
 -s, --strategy     git-duet, git-together (default: auto)
 -v, --verbose      verbose output (default: false)
```
  
### Supports
- [git-duet](https://github.com/git-duet/git-duet)
- [git-together](https://github.com/kejadlen/git-together)

  
### Installation
```brew install erictsiliacos/tap/portal```

## Contribute

### Bats
Bats is a bash testing framework, used here for integration tests. This can be installed with homebrew.

```brew bundle```

### Testing
```go test ./...```

```bats ./bats```

### Credit 

George Dean (@gdean123) for the initial idea
