setup_portal() {
  go build -o "$BATS_TMPDIR"/bin/portal
  PATH=$BATS_TMPDIR/bin:$PATH
}

init() {
  pushd "$1" || exit

  echo -ne '\n' | portal init %2>null

  git add .
  git commit -m "Add .portal files"
  git push origin master

  popd || exit
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
