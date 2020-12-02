package constants

import "fmt"

const EMPTY_INDEX = "nothing to push!"
const PORTAL_CLOSED = "nothing to pull!"
const REMOTE_TRACKING_REQUIRED = "must be on a branch that is remotely tracked"

func LOCAL_BRANCH_EXISTS(branch string) string {
	return fmt.Sprintf("local branch %s already exists", branch)
}

func REMOTE_BRANCH_EXISTS(branch string) string {
	return fmt.Sprintf("remote branch %s already exists", branch)
}

func DIRTY_INDEX(branch string) string {
	return fmt.Sprintf("%s: git index dirty!", branch)
}

func BRANCH_MISMATCH(startingBranch string, workingBranch string) string {
	return fmt.Sprintf("Starting branch %s did not match target branch %s", startingBranch, workingBranch)
}
