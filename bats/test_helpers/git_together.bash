brew_install_git_together() {
  git-together || brew install pivotal/tap/git-together
}

add_git_together() {
  pushd "$1" || exit
  git config --file .git-together --add git-together.domain email.com
  git config --file .git-together --add git-together.authors.fp 'Fake Person; fperson'
  git config --file .git-together --add git-together.authors.op 'Other Person; operson'
  git add .
  git commit -am "Add .git-together"
  git push origin main
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
