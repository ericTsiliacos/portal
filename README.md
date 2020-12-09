# 🌀 Portal

## Usage

Pushes local changes to a remote branch based on your pairing.

```bash
portal push
```

Pulls the changes at the pair branch name and cleans up the temporary branch

```bash
portal pull
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
