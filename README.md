# 🌀 Portal

### Installation
```brew install erictsiliacos/tap/portal```

### Usage
Assumsing both pairs have done `git duet <person1> <person2>` or `git-together with <person1> <person2>`

- `portal push`: pushes local changes to a remote branch based on your pairing (ex. portal-pa-ir)

- `portal pull`: pulls the changes at the pair branch name and cleans up the temporary branch
  
### Assumes
- You have git installed
- Either [git-duet](https://github.com/git-duet/git-duet) or [git-together](https://github.com/kejadlen/git-together) usage
- Both you and your pair have write access to the working repository
  
### Bats
Bats is a bash testing framework, used here for integration tests. This can be installed with homebrew.

```brew install bats```

### Testing

```bats ./bats```

### Credit 

George Dean (@gdean123) for the initial idea
