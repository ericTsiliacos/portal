package constants

import "fmt"

const EmptyIndex = "nothing to push!"
const PortalClosed = "nothing to pull!"
const RemoteTrackingRequired = "must be on a branch that is remotely tracked"
const DifferentVersions = `
Pusher and Puller are using different versions of portal
 1. Pusher run portal pull to retrieve changes.
 2. Both pairs update to latest version of portal.
Then try again...`
const GitProject = "not a git project"

func LocalBranchExists(branch string) string {
	return fmt.Sprintf("local branch %s already exists", branch)
}

func RemoteBranchExists(branch string) string {
	return fmt.Sprintf("remote branch %s already exists", branch)
}

func DirtyIndex(branch string) string {
	return fmt.Sprintf("%s: git index dirty!", branch)
}

func BranchMismatch(startingBranch string, workingBranch string) string {
	return fmt.Sprintf("Starting branch %s did not match target branch %s", startingBranch, workingBranch)
}
