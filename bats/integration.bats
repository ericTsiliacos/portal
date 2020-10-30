#!/usr/bin/env bats

@test "git-duet: push/pull" {
  add_git_duet "clone1" "clone2"

  git_duet "clone1"
  portal_push "clone1"

  git_duet "clone2"
  portal_pull "clone2"

  portal_push "clone2"
  portal_pull "clone1"
}

@test "git-together: push/pull" {
  add_git_together "clone1" "clone2"

  git_together "clone1"
  portal_push "clone1"

  git_together "clone2"
  portal_pull "clone2"

  portal_push "clone2"
  portal_pull "clone1"
}

@test "validate portal push found single branch naming strategy" {
  add_git_duet "clone1" "clone2"
  add_git_together "clone1" "clone2"

  git_together "clone1"
  git_duet "clone1"

  pushd "clone1" || exit

  touch foo.text

  run portal push
  [ "$status" -eq 1 ]
  [ "$output" = "Error: multiple branch naming strategies found" ]

  popd || exit
}

@test "validate portal pull found single branch naming strategy" {
  add_git_duet "clone1" "clone2"
  add_git_together "clone1" "clone2"

  git_together "clone1"
  git_duet "clone1"

  pushd "clone1" || exit

  run portal pull
  [ "$status" -eq 1 ]
  [ "$output" = "Error: multiple branch naming strategies found" ]

  popd || exit
}

@test "validate clean index before pulling" {
  cd clone1
  touch foo.text
  run git status --porcelain=v1
  [ "$output" = "?? foo.text" ]

  run portal pull
  [ "$output" = "git index dirty!" ]
}

@test "validate nonexistent remote branch before pushing" {
  add_git_duet "clone1" "clone2"

  git_duet "clone1"

  cd clone1
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
  brew_install_git_duet
  brew_install_git_together
  go_build_portal
  clean_test
  git_init_bare "project"
  git_clone "project" "clone1"
  git_clone "project" "clone2"
}

clean_bin() {
  rm -rf "${BATS_TMPDIR:?BATS_TMPDIR not set}"/bin
}

brew_install_git_duet() {
  git-duet || brew install git-duet/tap/git-duet
}

brew_install_git_together() {
  git-together || brew install pivotal/tap/git-together
}

go_build_portal() {
  go build -o "$BATS_TMPDIR"/bin/portal
  PATH=$BATS_TMPDIR/bin:$PATH
}

clean_test() {
  rm -rf "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}"
  mkdir -p "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}"
  cd "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}" || exit
}

git_init_bare() {
  mkdir "$1" && pushd "$1" && git init --bare && popd || exit
}

git_clone() {
  git clone "$1" "$2"
  pushd "$2" || exit
  git config user.name test
  git config user.email test@local
  popd || exit
}

add_git_duet() {
  pushd "$1" || exit
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

  pushd "$2" || exit
  git pull -r
  popd || exit
}

git_duet() {
  pushd "$1" || exit

  git-duet fp op

  popd || exit
}

add_git_together() {
  pushd "$1" || exit
  git config --file .git-together --add git-together.domain email.com
  git config --file .git-together --add git-together.authors.fp 'Fake Person; fperson'
  git config --file .git-together --add git-together.authors.op 'Other Person; operson'
  git add .
  git commit -am "Add .git-together"
  git push origin master
  popd || exit

  pushd "$2" || exit
  git pull -r
  popd || exit
}

git_together() {
  pushd "$1" || exit

  git-together with fp op

  popd || exit
}

portal_push() {
  pushd "$1" || exit

  touch foo.text

  run portal push
  [ "$status" -eq 0 ]

  run git status --porcelain=v1
  [ "$output" = "" ]

  popd || exit
}

portal_pull() {
  pushd "$1" || exit

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
