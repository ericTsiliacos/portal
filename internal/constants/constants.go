package constants

import "fmt"

const EMPTY_INDEX = "nothing to push!"
const PORTAL_CLOSED = "nothing to pull!"
const REMOTE_TRACKING_REQUIRED = "must be on a branch that is remotely tracked"
const DIFFERENT_VERSIONS = `
Pusher and Puller are using different versions of portal
 1. Pusher run portal pull to retrieve changes.
 2. Both pairs update to latest version of portal.
Then try again...`

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
