#!/usr/bin/env bats

load '/usr/local/lib/bats-support/load.bash'
load '/usr/local/lib/bats-assert/load.bash'
load './test_helpers/setup.bash'
load './test_helpers/git_repository.bash'
load './test_helpers/git_duet.bash'
load './test_helpers/git_together.bash'
load './test_helpers/portal.bash'

@test "git-duet: push/pull" {
  skip

  add_git_duet "clone1" "clone2"
  git_duet "clone1"
  git_duet "clone2"

  portal_push "clone1"
  portal_pull "clone2"

  portal_push "clone2"
  portal_pull "clone1"
}

@test "git-together: push/pull" {
  skip

  add_git_together "clone1" "clone2"
  git_together "clone1"
  git_together "clone2"

  portal_push "clone1"
  portal_pull "clone2"

  portal_push "clone2"
  portal_pull "clone1"
}

@test "push/pull: only commits and resets work-in-progress leaving nothing behind" {
  add_git_together "clone1" "clone2"
  git_together "clone1"
  git_together "clone2"

  pushd clone1
  touch foo.text
  git add .
  git commit -m "work in progress"
  run test_portal push
  assert_success

  run git cherry -v
  assert_output ""
  popd

  pushd clone2
  run test_portal pull --verbose
  assert_success

  run git cherry -v
  assert_output -p "work in progress"
}

@test "push: stashes changes for safe-keeping" {
  skip

  add_git_together "clone1" "clone2"

  git_together "clone1"

  pushd clone1
  touch foobar.text
  run test_portal push
  assert_success

  run git stash list -n 1
  assert_output -p "portal"
}

pull_validation() {
  @test "pull: validate found single branch naming strategy" {
    skip

    add_git_duet "clone1" "clone2"
    add_git_together "clone1" "clone2"

    git_together "clone1"
    git_duet "clone1"

    pushd "clone1" || exit

    run test_portal pull

    assert_failure
    assert_output "Error: multiple branch naming strategies found"
  }

  @test "pull: validate clean index" {
    skip

    cd clone1
    touch foo.text

    run test_portal pull

    assert_failure
    assert_output "git index dirty!"
  }

  @test "pull: validate existent remote branch" {
    skip

    add_git_duet "clone1" "clone2"
    git_duet "clone1"
    git_duet "clone2"

    cd clone2
    run test_portal pull

    assert_failure
    assert_output "remote branch portal-fp-op does not exists"
  }

  @test "pull: validate no unpublished work" {
    skip

    cd clone1
    touch foo.text
    git add .
    git commit -m "Un-pushed work"

    run test_portal pull -v

    assert_failure
    assert_output "Unpublished work detected."
  }
}

push_validation() {
  @test "push: validate found single branch naming strategy" {
    skip

    add_git_duet "clone1" "clone2"
    add_git_together "clone1" "clone2"

    git_together "clone1"
    git_duet "clone1"

    pushd "clone1" || exit

    touch foo.text

    run test_portal push

    assert_failure
    assert_output "Error: multiple branch naming strategies found"
  }

  @test "push: validate nonexistent remote branch" {
    skip

    add_git_duet "clone1" "clone2"

    git_duet "clone1"

    cd clone1
    touch foo.text
    git checkout -b portal-fp-op
    git add .
    git commit -m "WIP"
    git push -u origin portal-fp-op

    run test_portal push

    assert_failure
    assert_output "remote branch portal-fp-op already exists"
  }
}

pull_validation
push_validation

setup() {
  setup_file

  add_bin_to_path
  create_test
  git_init_bare "project"
  git_clone "project" "clone1"
  git_clone "project" "clone2"
}

setup_file() {
 if [[ "$BATS_TEST_NUMBER" -eq 1 ]]; then
    brew_install_git_duet
    brew_install_git_together
    go_build_portal
  fi
}

teardown() {
  clean_test

  if [[ "${#BATS_TEST_NAMES[@]}" -eq "$BATS_TEST_NUMBER" ]]; then
    clean_bin
  fi
}
