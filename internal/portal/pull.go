package portal

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/saga"
)

func PullSagaSteps(ctx context.Context, startingBranch string, portalBranch string, pusherSha string, verbose bool) (steps []saga.Step, err error) {
	remoteTrackingBranch, err := git.GetRemoteTrackingBranch()
	if err != nil {
		return
	}

	currentBranch, err := git.GetCurrentBranch()
	if err != nil {
		return
	}

	startingSha, err := git.GetBoundarySha(remoteTrackingBranch, currentBranch)
	if err != nil {
		return
	}

	return []saga.Step{
		{
			Name: "git rebase against remote working branch",
			Run: func() (err error) {
				return Run(exec.CommandContext(ctx, "git", "rebase", fmt.Sprintf("origin/%s", startingBranch)), verbose)
			},
			Undo: func() (err error) {
				return Run(exec.Command("git", "reset", "--hard", startingSha), verbose)
			},
		},
		{
			Name: "git reset to pusher sha",
			Run: func() (err error) {
				return Run(exec.CommandContext(ctx, "git", "reset", "--hard", pusherSha), verbose)
			},
		},
		{
			Name: "git rebase portal work in progress",
			Run: func() (err error) {
				return Run(exec.CommandContext(ctx, "git", "rebase", fmt.Sprintf("origin/%s", portalBranch)), verbose)
			},
		},
		{
			Name: "git reset commits",
			Run: func() (err error) {
				return Run(exec.CommandContext(ctx, "git", "reset", "HEAD^"), verbose)
			},
			Undo: func() (err error) {
				return Run(exec.Command("git", "add", "."), verbose)
			},
		},
		{
			Name: "delete remote portal branch",
			Run: func() (err error) {
				return Run(exec.CommandContext(ctx, "git", "push", "origin", "--delete", portalBranch, "--progress"), verbose)
			},
		},
	}, nil
}
