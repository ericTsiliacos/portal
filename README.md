# 🌀 Portal

# Usage

> You can use the `--help` option to get more details about the commands and their options

## Getting Started

Set up the same pairing using [git-duet](https://github.com/git-duet/git-duet) or [git-together](https://github.com/kejadlen/git-together)

## Push

Push local changes to a remote branch based on your pairing

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

### Patch option

Creates and stashes a git patch of the changes to be pushed.

The patch can be found at the head of `git stash list`. 

To apply the patch run: `git am <patch>`.

```bash
portal push --patch
```

## Pull

Pull changes at the pair branch name and clean up the temporary branch

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
