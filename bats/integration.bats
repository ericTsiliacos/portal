#!/usr/bin/env bats

setup() {
  rm -rf "${BATS_TMPDIR:?BATS_TMPDIR not set}"/bin
  go build -o "$BATS_TMPDIR"
  PATH=$BATS_TMPDIR/bin:$PATH

  rm -rf "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}"
  mkdir -p "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}"
  cd "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}" || exit

  cp -r "$BATS_TEST_DIRNAME"/project "$BATS_TMPDIR"/"$BATS_TEST_NAME"
  git clone project clone1
  git clone project clone2
}

@test "checking for dirty index before pulling" {
  cd clone1
  touch foo.text
  run git status --porcelain=v1
  [ "$output" = "?? foo.text" ]

  run portal pull
  [ "$output" = "git index dirty!" ]
}