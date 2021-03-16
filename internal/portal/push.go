package portal

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/saga"
	"gopkg.in/yaml.v2"
)

func PushSagaSteps(ctx context.Context, portalBranch string, now string, version string) []saga.Step {
	remoteTrackingBranch := git.GetRemoteTrackingBranch()
	currentBranch := git.GetCurrentBranch()
	sha := git.GetBoundarySha(remoteTrackingBranch, currentBranch)

	return []saga.Step{
		{
			Name: "git add .",
			Run: func() (err error) {
				cmd := exec.CommandContext(ctx, "git", "add", ".")
				log.Println(cmd.String())
				err = cmd.Run()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "reset")
				log.Println(cmd.String())
				err = cmd.Run()
				return
			},
		},
		{
			Name: "git commit -m 'portal-wip'",
			Run: func() (err error) {
				c := Meta{}
				c.Meta.WorkingBranch = currentBranch
				c.Meta.Sha = sha
				c.Meta.Version = version

				data, marshalError := yaml.Marshal(&c)
				if marshalError != nil {
					return err
				}

				cmd := exec.CommandContext(ctx, "git", "commit", "--allow-empty", "-m", string(data))
				log.Println(cmd.String())
				err = cmd.Run()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "reset", "HEAD^")
				log.Println(cmd.String())
				err = cmd.Run()
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

				cmd := exec.CommandContext(ctx, "git", "format-patch", remoteTrackingBranch, "--stdout")
				cmd.Stdout = file
				log.Println(cmd.String())
				err = cmd.Run()

				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("rm", BuildPatchFileName(now))
				log.Println(cmd.String())
				err = cmd.Run()
				return
			},
		},
		{
			Name: "git stash backup patch",
			Run: func() (err error) {
				cmd := exec.CommandContext(ctx, "git", "stash", "push", "-m", fmt.Sprintf("\"portal-patch-%s\"", now), "--include-untracked")
				log.Println(cmd.String())
				err = cmd.Run()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "stash", "pop")
				log.Println(cmd.String())
				err = cmd.Run()
				return
			},
		},
		{
			Name: "git checkout portal branch",
			Run: func() (err error) {
				cmd := exec.CommandContext(ctx, "git", "checkout", "-b", portalBranch, "--progress")
				log.Println(cmd.String())
				err = cmd.Run()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "checkout", currentBranch, "--progress")
				log.Println(cmd.String())
				if err = cmd.Run(); err != nil {
					return
				}

				cmd = exec.Command("git", "branch", "-D", portalBranch)
				log.Println(cmd.String())
				if err = cmd.Run(); err != nil {
					return
				}

				return
			},
		},
		{
			Name: "git push portal branch",
			Run: func() (err error) {
				cmd := exec.CommandContext(ctx, "git", "push", "origin", portalBranch, "--progress")
				log.Println(cmd.String())
				output, err := cmd.CombinedOutput()
				log.Println(err.Error())

				if err != nil {
					return err
				}

				log.Println(string(output))

				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "push", "origin", "--delete", portalBranch, "--progress")
				log.Println(cmd.String())
				err = cmd.Run()
				return
			},
			Retries: 0,
		},
		{
			Name: "git checkout to original branch",
			Run: func() (err error) {
				cmd := exec.CommandContext(ctx, "git", "checkout", "-", "--progress")
				log.Println(cmd.String())
				err = cmd.Run()
				return
			},
		},
		{
			Name: "delete local portal branch",
			Run: func() (err error) {
				cmd := exec.CommandContext(ctx, "git", "branch", "-D", portalBranch)
				log.Println(cmd.String())
				err = cmd.Run()
				return
			},
			Undo: func() (err error) {
				cmd := exec.Command("git", "checkout", "-b", portalBranch, "--progress")
				log.Println(cmd.String())
				err = cmd.Run()
				return
			},
		},
		{
			Name: "clear git workspace",
			Run: func() (err error) {
				cmd := exec.CommandContext(ctx, "git", "reset", "--hard", remoteTrackingBranch)
				log.Println(cmd.String())
				err = cmd.Run()
				return
			},
		},
	}
}
