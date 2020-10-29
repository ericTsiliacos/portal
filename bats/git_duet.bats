#!/usr/bin/env bats

@test "git-duet: push/pull" {
  push "clone1"
  pull "clone2"

  push "clone2"
  pull "clone1"
}

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

setup() {
  clean_bin
  install_git_duet
  install_portal
  clean_test
  create_remote_repo
  clone1 && add_git_duet
  clone2
}

install_git_duet() {
  pushd "$BATS_TMPDIR" || exit
  GOBIN="$BATS_TMPDIR"/bin go get github.com/git-duet/git-duet/...
  popd || exit
}

add_git_duet() {
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

install_portal() {
  go build -o "$BATS_TMPDIR"/bin/portal
  PATH=$BATS_TMPDIR/bin:$PATH
}

push() {
  pushd "$1" || exit

  run git-duet fp op
  [ "$status" -eq 0 ]

  touch foo.text

  run portal push
  [ "$status" -eq 0 ]

  run git status --porcelain=v1
  [ "$output" = "" ]

  popd || exit
}

pull() {
  pushd "$1" || exit

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

  popd || exit
}

create_remote_repo() {
  mkdir project && pushd project && git init --bare && popd || exit
}

clone1() {
  git clone project clone1
  pushd clone1 || exit
  git config user.name test
  git config user.email test@local
  popd || exit
}

clone2() {
  git clone project clone2
  pushd clone2 || exit
  git config user.name test
  git config user.email test@local
  git pull -r
  popd || exit
}

clean_test() {
  rm -rf "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}"
  mkdir -p "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}"
  cd "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}" || exit
}

clean_bin() {
  rm -rf "${BATS_TMPDIR:?BATS_TMPDIR not set}"/bin
}
