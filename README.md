# 🌀 Portal

Credit to George Dean (@gdean123) for the idea.

### Installation
```brew install erictsiliacos/tap/portal```

### Usage
Assumsing both pairs have done `git duet <person1> <person2>` or `git-together with <person1> <person2>`

- `portal push`: pushes local changes to a remote branch based on your pairing (ex. portal-pa-ir)

- `portal pull`: pulls the changes at the pair branch name and cleans up the temporary branch
  
### Assumes
- You have git installed
- Supports [git duet](https://github.com/git-duet/git-duet) and [git-together](https://github.com/kejadlen/git-together)
- Both you and your pair have write access
  
### Bats
Bats is a bash testing framework, used here for integration tests. This can be installed with homebrew.

```brew install bats```

### Testing

```bats ./bats```
