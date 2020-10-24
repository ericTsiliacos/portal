#!/usr/bin/env bats

@test "git-duet: push then pull example" {
  pushd clone1
  run git-duet fp op
  [ "$status" -eq 0 ]

  touch foo.text

  run portal push
  [ "$status" -eq 0 ]

  run git status --porcelain=v1
  [ "$output" = "" ]

  popd

  pushd clone2
  run git-duet fp op
  [ "$status" -eq 0 ]

  run git status --porcelain=v1
  [ "$output" = "" ]

  run portal pull
  echo "$output"
  [ "$status" -eq 0 ]

  run git status --porcelain=v1
  [ "$output" = "?? foo.text" ]

  run git ls-remote --heads origin portal-fp-op
  [ "$output" = "" ]
}

load "setup_project"
load "setup_portal"

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
