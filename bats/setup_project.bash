remoteRepo() {
  mkdir project && pushd project && git init --bare && popd || exit
}

clone1() {
  git clone project clone1
  pushd clone1 || exit
  git config user.name test
  git config user.email test@local
  popd || exit
}

clone2() {
  git clone project clone2
  pushd clone2 || exit
  git config user.name test
  git config user.email test@local
  git pull -r
  popd || exit
}

clean_test() {
  rm -rf "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}"
  mkdir -p "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}"
  cd "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}" || exit
}

clean_bin() {
  rm -rf "${BATS_TMPDIR:?BATS_TMPDIR not set}"/bin
}
