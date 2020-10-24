#!/usr/bin/env bats

@test "validate clean index before pulling" {
  cd clone1
  touch foo.text
  run git status --porcelain=v1
  [ "$output" = "?? foo.text" ]

  run portal pull
  [ "$output" = "git index dirty!" ]
}

@test "validate remote branch exists before pushing" {
  cd clone1
  git-duet fp op
  touch foo.text
  git checkout -b portal-fp-op
  git add .
  git commit -m "WIP"
  git push -u origin portal-fp-op

  run portal push
  [ "$output" = "remote branch portal-fp-op already exists" ]
}

load "helpers/repos"
load "helpers/portal"

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
