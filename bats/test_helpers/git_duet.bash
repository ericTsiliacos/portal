brew_install_git_duet() {
  git-duet || brew install git-duet/tap/git-duet
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
  git push origin main
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
