go_build_portal() {
  go build -ldflags "-X main.version=v1" -o "$BATS_TMPDIR"/bin/test_portal
  go build -ldflags "-X main.version=v2" -o "$BATS_TMPDIR"/bin/test_portal_v2
}

portal_push() {
  pushd "$1" || exit

  touch foo.text

  run test_portal push
  assert_success

  run git status --porcelain=v1
  assert_output ""

  popd || exit
}

portal_pull() {
  pushd "$1" || exit

  run git status --porcelain=v1
  assert_output ""

  run test_portal pull
  assert_success

  run git status --porcelain=v1
  assert_output "?? foo.text"

  run git ls-remote --heads origin portal-fp-op
  assert_output ""

  popd || exit
}
