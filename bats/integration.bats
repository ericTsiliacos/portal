#!/usr/bin/env bats

setup() {
  rm -rf "${BATS_TMPDIR:?BATS_TMPDIR not set}"/bin

  pushd "$BATS_TMPDIR" || exit
  GOBIN="$BATS_TMPDIR"/bin go get github.com/git-duet/git-duet/...
  popd || exit

  go build -o "$BATS_TMPDIR"/bin/portal
  PATH=$BATS_TMPDIR/bin:$PATH

  rm -rf "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}"
  mkdir -p "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}"
  cd "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}" || exit

  cp -r "$BATS_TEST_DIRNAME"/project "$BATS_TMPDIR"/"$BATS_TEST_NAME"
  git clone project clone1
  git clone project clone2
}

@test "push then pull example" {
  pushd clone1
  git-duet fp op
  touch foo.text

  run portal push
  [ "$status" -eq 0 ]

  run git status --porcelain=v1
  [ "$output" = "" ]

  popd

  pushd clone2
  git-duet fp op

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

@test "checks for dirty index before pulling" {
  cd clone1
  touch foo.text
  run git status --porcelain=v1
  [ "$output" = "?? foo.text" ]

  run portal pull
  [ "$output" = "git index dirty!" ]
}

@test "check for existing remote branch before pushing" {
  cd clone1
  git-duet fp op
  touch foo.text
  git checkout -b portal-fp-op
  git add .
  git commit -m "WIP"
  git push origin head

  run portal push
  [ "$output" = "remote branch portal-fp-op already exists" ]
}
