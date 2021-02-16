package portal

import (
	"fmt"

	"github.com/ericTsiliacos/portal/internal/saga"
	"github.com/ericTsiliacos/portal/internal/shell"
)

func PullSagaSteps(startingBranch string, portalBranch string, pusherSha string) []saga.Step {
	return []saga.Step{
		{
			Name: "git rebase against remote working branch",
			Run: func() (err error) {
				_, err = shell.Execute(fmt.Sprintf("git rebase origin/%s", startingBranch))
				return
			},
			Undo: func() (err error) {
				return
			},
		},
		{
			Name: "git reset to pusher sha",
			Run: func() (err error) {
				_, err = shell.Execute(fmt.Sprintf("git reset --hard %s", pusherSha))
				return
			},
			Undo: func() (err error) {
				return
			},
		},
		{
			Name: "git rebase portal work in progress",
			Run: func() (err error) {
				_, err = shell.Execute(fmt.Sprintf("git rebase origin/%s~1", portalBranch))
				return
			},
			Undo: func() (err error) {
				return
			},
		},
		{
			Name: "git reset commits",
			Run: func() (err error) {
				_, err = shell.Execute("git reset HEAD^")
				return
			},
			Undo: func() (err error) {
				return
			},
		},
		{
			Name: "delete remote portal branch",
			Run: func() (err error) {
				_, err = shell.Execute(fmt.Sprintf("git push origin --delete %s --progress", portalBranch))
				return
			},
			Undo: func() (err error) {
				return
			},
		},
	}
}
