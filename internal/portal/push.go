package portal

import (
	"fmt"

	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/saga"
	"github.com/ericTsiliacos/portal/internal/shell"
)

func PushSagaSteps(portalBranch string, now string, version string, patch bool) []saga.Step {
	remoteTrackingBranch := git.GetRemoteTrackingBranch()
	currentBranch := git.GetCurrentBranch()
	sha := git.GetBoundarySha(remoteTrackingBranch, currentBranch)

	return []saga.Step{
		{
			Name:       "git add .",
			Run:        func() (string, error) { return git.Add(".") },
			Compensate: func(input string) (string, error) { return git.Reset() },
		},
		{
			Name:       "git commit -m 'portal-wip'",
			Run:        func() (string, error) { return git.Commit("portal-wip") },
			Compensate: func(input string) (string, error) { return git.UndoCommit() },
		},
		{
			Name: "create git patch of work in progress",
			Run: func() (string, error) {
				return Patch(remoteTrackingBranch, now)
			},
			Compensate: func(fileName string) (string, error) {
				return shell.Execute(fmt.Sprintf("rm %s", fileName))
			},
			Exclude: !patch,
		},
		{
			Name: "create portal-meta.yml",
			Run: func() (string, error) {
				return WritePortalMetadata("portal-meta.yml", currentBranch, sha, version)
			},
			Compensate: func(fileName string) (string, error) {
				return shell.Execute(fmt.Sprintf("rm %s", fileName))
			},
		},
		{
			Name:       "git add portal-meta.yml",
			Run:        func() (string, error) { return git.Add("portal-meta.yml") },
			Compensate: func(input string) (string, error) { return git.Reset() },
		},
		{
			Name:       "git commit -m 'portal-meta'",
			Run:        func() (string, error) { return git.Commit("portal-meta") },
			Compensate: func(input string) (string, error) { return git.UndoCommit() },
		},
		{
			Name: "git stash backup patch",
			Run: func() (string, error) {
				return shell.Execute(fmt.Sprintf("git stash push -m \"portal-patch-%s\" --include-untracked", now))
			},
			Compensate: func(input string) (string, error) {
				return shell.Execute("git stash pop")
			},
			Exclude: !patch,
		},
		{
			Name: "git checkout portal branch",
			Run: func() (string, error) {
				_, err := shell.Execute(fmt.Sprintf("git checkout -b %s --progress", portalBranch))
				return currentBranch, err
			},
			Compensate: func(currentBranch string) (string, error) {
				shell.Execute(fmt.Sprintf("git checkout %s --progress", currentBranch))

				return shell.Execute(fmt.Sprintf("git branch -D %s", portalBranch))
			},
		},
		{
			Name: "git push portal branch",
			Run: func() (string, error) {
				return shell.Execute(fmt.Sprintf("git push origin %s --progress", portalBranch))
			},
			Compensate: func(originalBranch string) (string, error) {
				return shell.Execute(fmt.Sprintf("git push origin --delete %s --progress", portalBranch))
			},
			Retries: 1,
		},
		{
			Name: "git checkout to original branch",
			Run: func() (string, error) {
				return shell.Execute("git checkout - --progress")
			},
		},
		{
			Name: "delete local portal branch",
			Run: func() (string, error) {
				return shell.Execute(fmt.Sprintf("git branch -D %s", portalBranch))
			},
			Compensate: func(originalBranch string) (string, error) {
				return shell.Execute(fmt.Sprintf("git checkout -b %s --progress", portalBranch))
			},
		},
		{
			Name: "clear git workspace",
			Run: func() (string, error) {
				return shell.Execute(fmt.Sprintf("git reset --hard %s", remoteTrackingBranch))
			},
		},
	}
}
