package portal

import (
	"fmt"
	"os/exec"

	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/saga"
	"github.com/ericTsiliacos/portal/internal/shell"
)

func PushSagaSteps(portalBranch string, now string, version string) []saga.Step {
	remoteTrackingBranch := git.GetRemoteTrackingBranch()
	currentBranch := git.GetCurrentBranch()
	sha := git.GetBoundarySha(remoteTrackingBranch, currentBranch)

	return []saga.Step{
		{
			Name: "git add .",
			Run: func() (err error) {
				_, err = exec.Command("git", "add", ".").CombinedOutput()
				return
			},
			Undo: func() (err error) { _, err = git.Reset(); return },
		},
		{
			Name: "git commit -m 'portal-wip'",
			Run:  func() (err error) { _, err = git.Commit("portal-wip"); return },
			Undo: func() (err error) { _, err = git.UndoCommit(); return },
		},
		{
			Name: "create git patch of work in progress",
			Run: func() (err error) {
				_, err = Patch(remoteTrackingBranch, now)
				return
			},
			Undo: func() (err error) {
				_, err = shell.Execute(fmt.Sprintf("rm %s", BuildPatchFileName(now)))
				return
			},
		},
		{
			Name: "create portal-meta.yml",
			Run: func() (err error) {
				_, err = WritePortalMetadata("portal-meta.yml", currentBranch, sha, version)
				return
			},
			Undo: func() (err error) {
				_, err = shell.Execute("rm portal-meta.yml")
				return
			},
		},
		{
			Name: "git add portal-meta.yml",
			Run:  func() (err error) { _, err = git.Add("portal-meta.yml"); return },
			Undo: func() (err error) { _, err = git.Reset(); return },
		},
		{
			Name: "git commit -m 'portal-meta'",
			Run:  func() (err error) { _, err = git.Commit("portal-meta"); return },
			Undo: func() (err error) { _, err = git.UndoCommit(); return },
		},
		{
			Name: "git stash backup patch",
			Run: func() (err error) {
				_, err = shell.Execute(fmt.Sprintf("git stash push -m \"portal-patch-%s\" --include-untracked", now))
				return
			},
			Undo: func() (err error) {
				_, err = shell.Execute("git stash pop")
				return
			},
		},
		{
			Name: "git checkout portal branch",
			Run: func() (err error) {
				_, err = shell.Execute(fmt.Sprintf("git checkout -b %s --progress", portalBranch))
				return
			},
			Undo: func() (err error) {
				if _, err = shell.Execute(fmt.Sprintf("git checkout %s --progress", currentBranch)); err != nil {
					return
				}
				_, err = shell.Execute(fmt.Sprintf("git branch -D %s", portalBranch))
				return
			},
		},
		{
			Name: "git push portal branch",
			Run: func() (err error) {
				_, err = shell.Execute(fmt.Sprintf("git push origin %s --progress", portalBranch))
				return
			},
			Undo: func() (err error) {
				_, err = shell.Execute(fmt.Sprintf("git push origin --delete %s --progress", portalBranch))
				return
			},
			Retries: 1,
		},
		{
			Name: "git checkout to original branch",
			Run: func() (err error) {
				_, err = shell.Execute("git checkout - --progress")
				return
			},
		},
		{
			Name: "delete local portal branch",
			Run: func() (err error) {
				_, err = shell.Execute(fmt.Sprintf("git branch -D %s", portalBranch))
				return
			},
			Undo: func() (err error) {
				_, err = shell.Execute(fmt.Sprintf("git checkout -b %s --progress", portalBranch))
				return
			},
		},
		{
			Name: "clear git workspace",
			Run: func() (err error) {
				_, err = shell.Execute(fmt.Sprintf("git reset --hard %s", remoteTrackingBranch))
				return
			},
		},
	}
}
