git_init_bare() {
  mkdir "$1" && pushd "$1" && git init --bare && popd || exit

  git clone "$1" setup_project
  pushd setup_project || exit
  git commit --allow-empty -m "initial commit"
  git push origin head
  popd || exit
  rm -rf setup_project
}

git_clone() {
  git clone "$1" "$2"
  pushd "$2" || exit
  git config user.name test
  git config user.email test@local
  popd || exit
}
