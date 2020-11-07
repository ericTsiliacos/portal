add_bin_to_path() {
  PATH=$BATS_TMPDIR/bin:$PATH
}

clean_bin() {
  rm -rf "${BATS_TMPDIR:?BATS_TMPDIR not set}"/bin
}

create_test() {
  mkdir -p "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}"
  cd "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}" || exit
}

clean_test() {
  rm -rf "${BATS_TMPDIR:?}"/"${BATS_TEST_NAME:?}"
}
