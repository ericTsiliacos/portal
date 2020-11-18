#!/usr/bin/env bats

load '/usr/local/lib/bats-support/load.bash'
load '/usr/local/lib/bats-assert/load.bash'
load '/usr/local/lib/bats-file/load.bash'
load './test_helpers/setup.bash'
load './test_helpers/git_repository.bash'
load './test_helpers/git_duet.bash'
load './test_helpers/git_together.bash'
load './test_helpers/portal.bash'

@test "git-duet: push/pull" {
  add_git_duet "clone1" "clone2"
  git_duet "clone1"
  git_duet "clone2"

  portal_push "clone1"
  portal_pull "clone2"

  portal_push "clone2"
  portal_pull "clone1"
}

@test "git-together: push/pull" {
  add_git_together "clone1" "clone2"
  git_together "clone1"
  git_together "clone2"

  portal_push "clone1"
  portal_pull "clone2"

  portal_push "clone2"
  portal_pull "clone1"
}

@test "push/pull: commits and resets index" {
  add_git_together "clone1" "clone2"
  git_together "clone1"
  git_together "clone2"

  pushd clone1
  touch foo.text
  git add .
  git commit -m "work in progress"
  touch bar.text
  run test_portal push
  assert_success

  run git cherry -v
  assert_output ""
  popd

  pushd clone2
  run test_portal pull
  assert_success

  run git cherry -v
  assert_output -p "work in progress"
  assert_file_exist bar.text

}

@test "push/pull: when puller is ahead of pusher against origin" {
  add_git_duet "clone1" "clone2"
  git_duet "clone1"
  git_duet "clone2"

  git_clone "project" "clone3"
  pushd clone3
  touch something.text
  git add .
  git push origin head
  popd

  pushd clone2
  git pull -r
  popd

  portal_push "clone1"
  portal_pull "clone2"
}

@test "push/pull: when pusher is ahead of puller against origin" {
  add_git_duet "clone1" "clone2"
  git_duet "clone1"
  git_duet "clone2"

  git_clone "project" "clone3"
  pushd clone3
  touch something.text
  git add .
  git push origin head
  popd

  pushd clone1
  git pull -r
  popd

  portal_push "clone1"
  portal_pull "clone2"
}

@test "push: stashes a patch of changes for safe-keeping" {
  add_git_together "clone1" "clone2"

  git_together "clone1"

  pushd clone1
  touch foobar.text
  git add .
  git commit -m "work in progress"
  touch bar.text
  run test_portal push
  assert_success

  run git stash list -n 1
  assert_output -p "portal-patch-"

  git stash pop
  git am *.patch

  git reset HEAD^

  run git cherry -v
  assert_output -p "work in progress"

  assert_file_exist bar.text
}

@test "push/pull: option to provide branch naming strategy" {
  add_git_duet "clone1" "clone2"
  add_git_together "clone1" "clone2"

  git_together "clone1"
  git_duet "clone1"

  pushd "clone1" || exit

  touch foo.text

  run test_portal push

  assert_failure
  assert_output "Error: multiple branch naming strategies found"

  run test_portal push -s git-duet
  assert_success
  popd

  pushd "clone1" || exit

  run test_portal pull
  assert_failure
  assert_output "Error: multiple branch naming strategies found"

  run test_portal pull -s git-duet
  assert_success
}

@test "push/pull: validate branch naming strategy option" {
  add_git_duet "clone1" "clone2"
  add_git_together "clone1" "clone2"

  git_together "clone1"
  git_duet "clone1"

  pushd "clone1" || exit

  touch foo.text

  run test_portal push --strategy wrong
  assert_failure
  assert_output "Error: unknown strategy"
}

push_validation() {
  @test "push: validate dirty workspace" {
    cd clone1
    run test_portal push

    assert_failure
    assert_output "nothing to push!"
  }

  @test "push: validate branch is remotely traced" {
    add_git_duet "clone1" "clone2"
    git_duet "clone1"

    cd clone1
    git checkout -b some_branch
    touch foo.text

    run test_portal push
    assert_failure
    assert_output "Only branches with remote tracking are pushable"
  }

  @test "push: validate found single branch naming strategy" {
    add_git_duet "clone1" "clone2"
    add_git_together "clone1" "clone2"

    git_together "clone1"
    git_duet "clone1"

    pushd "clone1" || exit

    touch foo.text

    run test_portal push

    assert_failure
    assert_output "Error: multiple branch naming strategies found"

    run test_portal push -s git-duet
    assert_success
  }

  @test "push: validate local portal branch nonexistent" {
    add_git_duet "clone1" "clone2"

    git_duet "clone1"

    cd clone1
    touch foo.text
    git checkout -b portal-fp-op
    git checkout master

    run test_portal push
    assert_failure
    assert_output "local branch portal-fp-op already exists"
  }

  @test "push: validate nonexistent remote branch" {
    add_git_duet "clone1" "clone2"

    git_duet "clone1"

    cd clone1
    touch foo.text
    git checkout -b portal-fp-op
    git add .
    git commit -m "WIP"
    git push -u origin portal-fp-op

    git checkout master
    git branch -D portal-fp-op
    touch bar.text

    run test_portal push

    assert_failure
    assert_output "remote branch portal-fp-op already exists"
  }
}

pull_validation() {
  @test "pull: validate current branch is remotely tracked" {
    add_git_duet "clone1" "clone2"
    git_duet "clone1"
    git_duet "clone2"

    cd clone2
    git checkout -b untracked_branch
    run test_portal pull
    assert_failure
    assert_output "Must be on a branch that is remotely tracked."
  }

  @test "pull: validate no present commits" {
    cd clone1
    touch foo.text
    git add .
    git commit -m "Un-pushed work"

    run test_portal pull

    assert_failure
    assert_output "master: git index dirty!"
  }

  @test "pull: validate clean index" {
    cd clone1
    touch foo.text

    run test_portal pull

    assert_failure
    assert_output "master: git index dirty!"
  }

  @test "pull: validate existent remote branch" {
    add_git_duet "clone1" "clone2"
    git_duet "clone1"
    git_duet "clone2"

    cd clone2
    run test_portal pull

    assert_failure
    assert_output "nothing to pull!"
  }

  @test "pull: validate found single branch naming strategy" {
    add_git_duet "clone1" "clone2"
    add_git_together "clone1" "clone2"

    git_together "clone1"
    git_duet "clone1"

    pushd "clone1" || exit

    run test_portal pull

    assert_failure
    assert_output "Error: multiple branch naming strategies found"
  }

  @test "pull: validate puller is using same portal version as pusher" {
    add_git_together "clone1" "clone2"
    git_together "clone1"
    git_together "clone2"

    pushd clone1
    touch foo.text
    git add .
    git commit -m "work in progress"
    run test_portal push
    assert_success
    popd

    cd clone2
    run test_portal_v2 pull

    assert_failure
    assert_line --partial "Pusher and Puller are using different versions of portal"
  }

  @test "pull: validate puller is on same branch as pusher" {
    add_git_duet "clone1" "clone2"
    git_duet "clone1"
    git_duet "clone2"

    pushd clone1
    touch foo.text
    git add .
    git commit -m "foo.text"
    run test_portal push
    assert_success
    popd

    pushd clone2
    git checkout -b another_branch
    touch bar.text
    git add .
    git commit -m "new branch"
    git push origin -u another_branch

    run test_portal pull
    assert_failure
    assert_output "Starting branch another_branch did not match target branch master"
  }
}

push_validation
pull_validation

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
