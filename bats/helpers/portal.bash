setup_portal() {
  go build -o "$BATS_TMPDIR"/bin/portal
  PATH=$BATS_TMPDIR/bin:$PATH
}
