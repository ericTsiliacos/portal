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
