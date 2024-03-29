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

@test "portal push/pull from different directory levels" {
  add_git_together "clone1" "clone2"
  git_together "clone1"
  git_together "clone2"

  pushd clone1

  mkdir tmp
  touch tmp/bar.text
  git add .
  git commit -m "create additional directory structure"
  git push origin head

  touch foo.text
  cd tmp
  run test_portal push
  assert_success
  run git status --porcelain=v1
  assert_output ""

  popd

  pushd clone2
  git pull -r
  cd tmp

  run test_portal pull
  assert_success

  run git status
  assert_output -p "../foo.text"
}

@test "support adding message to the portal commit message" {
  add_git_together "clone1" "clone2"
  git_together "clone1"
  git_together "clone2"

  pushd clone1 || exit

  touch foo.text

  PORTAL_COMMIT_MESSAGE="message goes here" run test_portal push
  assert_success

  run git log origin/tmp/portal/fp-op --format=%B -n 1

  assert_output -p "message goes here"
  popd || exit

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

@test "push/pull: when puller is ahead of pusher against remote" {
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

@test "push/pull: when pusher is ahead of puller against remote" {
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

@test "strategy option: not configured" {
  add_git_duet "clone1" "clone2"

  pushd "clone1" || exit
  touch foo.text
  run test_portal push -s git-duet
  assert_failure
  assert_output "Error: git-duet not configured"
  popd

  git_duet "clone1"

  pushd "clone1" || exit
  run test_portal push -s git-duet
  assert_success
  popd

  pushd "clone2" || exit
  run test_portal pull -s git-duet
  assert_failure
  assert_output "Error: git-duet not configured"
  popd
}

@test "strategy option: validate branch option" {
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
  @test "push: validate current working directory is inside a working git tree" {
    add_git_duet "clone1" "clone2"
    git_duet "clone1"
    git_duet "clone2"

    pushd clone1
    mkdir tmp
    cd tmp 
    touch bar.text
    run test_portal push
    assert_success
    popd

    mkdir non_git_project
    cd non_git_project
    run test_portal push

    assert_failure
    assert_output "not a git project"
  }

  @test "push: validate dirty workspace" {
    add_git_duet "clone1" "clone2"
    git_duet "clone1"
    cd clone1
    run test_portal push

    assert_failure
    assert_output "nothing to push!"
  }

  @test "push: validate branch is remotely tracked" {
    add_git_duet "clone1" "clone2"
    git_duet "clone1"

    cd clone1
    git checkout -b some_branch
    touch foo.text

    run test_portal push
    assert_failure
    assert_output "must be on a branch that is remotely tracked"
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
    git checkout -b tmp/portal/fp-op
    git checkout main

    run test_portal push
    assert_failure
    assert_output "local branch tmp/portal/fp-op already exists"
  }

  @test "push: validate nonexistent remote branch" {
    add_git_duet "clone1" "clone2"

    git_duet "clone1"

    cd clone1
    touch foo.text
    git checkout -b tmp/portal/fp-op
    git add .
    git commit -m "WIP"
    git push -u origin tmp/portal/fp-op

    git checkout main
    git branch -D tmp/portal/fp-op
    touch bar.text

    run test_portal push

    assert_failure
    assert_output "remote branch tmp/portal/fp-op already exists"
  }
}

pull_validation() {
   @test "pull: validate current working directory is inside a working git tree" {
    add_git_duet "clone1" "clone2"
    git_duet "clone1"
    git_duet "clone2"

    pushd clone1
    mkdir tmp
    cd tmp && touch bar.text && cd ..
    git add .
    git commit -m "directory added"
    git push origin head

    touch foo.text
    run test_portal push
    assert_success
    popd

    pushd clone2
    git pull --rebase
    cd tmp
    run test_portal pull
    assert_success
    popd

    mkdir non_git_project
    cd non_git_project
    run test_portal push

    assert_failure
    assert_output "not a git project"
  }

  @test "pull: validate current branch is remotely tracked" {
    add_git_duet "clone1" "clone2"
    git_duet "clone1"
    git_duet "clone2"

    pushd clone1 
    touch bar.text
    run test_portal push
    assert_success
    popd

    cd clone2
    git checkout -b untracked_branch
    run test_portal pull
    assert_failure
    assert_output "must be on a branch that is remotely tracked"
  }

  @test "pull: validate no present commits" {
    add_git_duet "clone1" "clone2"
    git_duet "clone1"
    git_duet "clone2"

    pushd clone1 
    touch bar.text
    run test_portal push
    assert_success
    popd

    cd clone2
    touch foo.text
    git add .
    git commit -m "unfinished work"
    run test_portal pull

    assert_failure
    assert_output "main: git index dirty!"
  }

  @test "pull: validate clean index" {
   add_git_duet "clone1" "clone2"
    git_duet "clone1"
    git_duet "clone2"

    pushd clone1 
    touch bar.text
    run test_portal push
    assert_success
    popd

    cd clone2
    touch foo.text
    run test_portal pull

    assert_failure
    assert_output "main: git index dirty!"
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
    assert_output "Starting branch another_branch did not match target branch main"
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
