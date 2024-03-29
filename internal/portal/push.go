package portal

import (
	"context"
	"os/exec"

	"gopkg.in/yaml.v2"

	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/saga"
	"github.com/ericTsiliacos/portal/internal/shell"
)

func PushSagaSteps(ctx context.Context, portalBranch string, version string, verbose bool, commitMessage string) (steps []saga.Step, err error) {
	remoteTrackingBranch, err := git.GetRemoteTrackingBranch()
	if err != nil {
		return
	}

	currentBranch, err := git.GetCurrentBranch()
	if err != nil {
		return
	}

	sha, err := git.GetBoundarySha(remoteTrackingBranch, currentBranch)
	if err != nil {
		return
	}

	return []saga.Step{
		{
			Name: "git add -A",
			Run: func() (err error) {
				return shell.Run(exec.CommandContext(ctx, "git", "add", "--all"), verbose)
			},
			Undo: func() (err error) {
				return shell.Run(exec.Command("git", "reset"), verbose)
			},
		},
		{
			Name: "git commit -m 'portal-wip'",
			Run: func() (err error) {
				config := Meta{}
				config.Meta.WorkingBranch = currentBranch
				config.Meta.Sha = sha
				config.Meta.Version = version
				config.Meta.Message = commitMessage

				data, marshalError := yaml.Marshal(&config)
				if marshalError != nil {
					return err
				}

				return shell.Run(exec.CommandContext(ctx, "git", "commit", "--allow-empty", "-m", string(data)), verbose)
			},
			Undo: func() (err error) {
				return shell.Run(exec.Command("git", "reset", "HEAD^"), verbose)
			},
		},
		{
			Name: "git checkout portal branch",
			Run: func() (err error) {
				return shell.Run(exec.CommandContext(ctx, "git", "checkout", "-b", portalBranch, "--progress"), verbose)
			},
			Undo: func() (err error) {
				if err = shell.Run(exec.Command("git", "checkout", currentBranch, "--progress"), verbose); err != nil {
					return
				}

				if err = shell.Run(exec.Command("git", "branch", "-D", portalBranch), verbose); err != nil {
					return
				}

				return
			},
		},
		{
			Name: "git push portal branch",
			Run: func() (err error) {
				return shell.Run(exec.CommandContext(ctx, "git", "push", "origin", portalBranch, "--progress"), verbose)
			},
			Undo: func() (err error) {
				return shell.Run(exec.Command("git", "push", "origin", "--delete", portalBranch, "--progress"), verbose)
			},
		},
		{
			Name: "git checkout to original branch",
			Run: func() (err error) {
				return shell.Run(exec.CommandContext(ctx, "git", "checkout", "-", "--progress"), verbose)
			},
		},
		{
			Name: "delete local portal branch",
			Run: func() (err error) {
				return shell.Run(exec.CommandContext(ctx, "git", "branch", "-D", portalBranch), verbose)
			},
			Undo: func() (err error) {
				return shell.Run(exec.Command("git", "checkout", "-b", portalBranch, "--progress"), verbose)
			},
		},
		{
			Name: "clear git workspace",
			Run: func() (err error) {
				return shell.Run(exec.CommandContext(ctx, "git", "reset", "--hard", remoteTrackingBranch), verbose)
			},
		},
	}, nil
}
