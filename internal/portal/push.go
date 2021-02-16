package portal

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/saga"
)

func PushSagaSteps(portalBranch string, now string, version string) []saga.Step {
	remoteTrackingBranch := git.GetRemoteTrackingBranch()
	currentBranch := git.GetCurrentBranch()
	sha := git.GetBoundarySha(remoteTrackingBranch, currentBranch)

	return []saga.Step{
		{
			Name: "git add .",
			Run: func() (err error) {
				cmd := exec.Command("git", "add", ".")
				_, err = cmd.CombinedOutput()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "reset")
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "git commit -m 'portal-wip'",
			Run: func() (err error) {
				cmd := exec.Command("git", "commit", "--allow-empty", "-m", "portal-wip")
				_, err = cmd.CombinedOutput()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "reset", "HEAD^")
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "create git patch of work in progress",
			Run: func() (err error) {
				patchFileName := BuildPatchFileName(now)
				file, err := os.Create(patchFileName)

				if err != nil {
					return err
				}

				defer file.Close()

				cmd := exec.Command("git", "format-patch", remoteTrackingBranch, "--stdout")
				cmd.Stdout = file
				err = cmd.Run()

				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("rm", BuildPatchFileName(now))
				_, err = cmd.CombinedOutput()
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
				cmd := exec.Command("rm", "portal-meta.yml")
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "git add portal-meta.yml",
			Run: func() (err error) {
				cmd := exec.Command("git", "add", "portal-meta.yml")
				_, err = cmd.CombinedOutput()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "reset")
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "git commit -m 'portal-meta'",
			Run: func() (err error) {
				cmd := exec.Command("git", "commit", "--allow-empty", "-m", "portal-meta")
				_, err = cmd.CombinedOutput()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "reset", "HEAD^")
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "git stash backup patch",
			Run: func() (err error) {
				cmd := exec.Command("git", "stash", "push", "-m", fmt.Sprintf("\"portal-patch-%s\"", now), "--include-untracked")
				_, err = cmd.CombinedOutput()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "stash", "pop")
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "git checkout portal branch",
			Run: func() (err error) {
				cmd := exec.Command("git", "checkout", "-b", portalBranch, "--progress")
				_, err = cmd.CombinedOutput()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "checkout", currentBranch, "--progress")
				if _, err = cmd.CombinedOutput(); err != nil {
					return
				}

				cmd = exec.Command("git", "branch", "-D", portalBranch)
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "git push portal branch",
			Run: func() (err error) {
				cmd := exec.Command("git", "push", "origin", portalBranch, "--progress")
				_, err = cmd.CombinedOutput()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "push", "origin", "--delete", portalBranch, "--progress")
				_, err = cmd.CombinedOutput()
				return
			},
			Retries: 1,
		},
		{
			Name: "git checkout to original branch",
			Run: func() (err error) {
				cmd := exec.Command("git", "checkout", "-", "--progress")
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "delete local portal branch",
			Run: func() (err error) {
				cmd := exec.Command("git", "branch", "-D", portalBranch)
				_, err = cmd.CombinedOutput()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "checkout", "-b", portalBranch, "--progress")
				_, err = cmd.CombinedOutput()
				return
			},
		},
		{
			Name: "clear git workspace",
			Run: func() (err error) {
				cmd := exec.Command("git", "reset", "--hard", remoteTrackingBranch)
				_, err = cmd.CombinedOutput()
				return
			},
		},
	}
}
