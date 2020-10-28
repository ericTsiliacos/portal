#!/usr/bin/env bats

@test "git-duet: push then pull example" {
  push "clone1"
  pull "clone2"

  push "clone2"
  pull "clone1"
}

load "helpers/repos.bash"
load "helpers/portal.bash"

setup() {
  clean_bin
  goGetGitDuet
  setup_portal
  clean_test
  setup_repos
}

setup_repos() {
  remoteRepo
  clone1 && addGitDuet
  clone2
}

goGetGitDuet() {
  pushd "$BATS_TMPDIR" || exit
  GOBIN="$BATS_TMPDIR"/bin go get github.com/git-duet/git-duet/...
  popd || exit
}

addGitDuet() {
  pushd clone1 || exit
  cat > .git-authors <<- EOM
authors:
  fp: Fake Person; fperson
  op: Other Person; operson
email_addresses:
  fp: fperson@email.com
  op: operson@email.com
EOM
  git add .
  git commit -am "Add .git-author"
  git push origin master
  popd || exit
}
