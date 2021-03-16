package portal

import (
	"fmt"
	"os/exec"

	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/saga"
)

func PullSagaSteps(startingBranch string, portalBranch string, pusherSha string) []saga.Step {
	remoteTrackingBranch := git.GetRemoteTrackingBranch()
	currentBranch := git.GetCurrentBranch()
	startingSha := git.GetBoundarySha(remoteTrackingBranch, currentBranch)

	return []saga.Step{
		{
			Name: "git rebase against remote working branch",
			Run: func() (err error) {
				cmd := exec.Command("git", "rebase", fmt.Sprintf("origin/%s", startingBranch))
				_, err = cmd.CombinedOutput()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "reset", "--hard", startingSha)
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "git reset to pusher sha",
			Run: func() (err error) {
				cmd := exec.Command("git", "reset", "--hard", pusherSha)
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "git rebase portal work in progress",
			Run: func() (err error) {
				cmd := exec.Command("git", "rebase", fmt.Sprintf("origin/%s", portalBranch))
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "git reset commits",
			Run: func() (err error) {
				cmd := exec.Command("git", "reset", "HEAD^")
				_, err = cmd.CombinedOutput()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "add", ".")
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "delete remote portal branch",
			Run: func() (err error) {
				cmd := exec.Command("git", "push", "origin", "--delete", portalBranch, "--progress")
				_, err = cmd.CombinedOutput()
				return
			},
		},
	}
}
