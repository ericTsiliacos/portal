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
			Run: func() (string, error) {
				return shell.Execute(fmt.Sprintf("git rebase origin/%s", startingBranch))
			},
			Compensate: func(input string) (string, error) {
				return "", nil
			},
		},
		{
			Name: "git reset to pusher sha",
			Run: func() (string, error) {
				return shell.Execute(fmt.Sprintf("git reset --hard %s", pusherSha))
			},
			Compensate: func(input string) (string, error) {
				return "", nil
			},
		},
		{
			Name: "git rebase portal work in progress",
			Run: func() (string, error) {
				return shell.Execute(fmt.Sprintf("git rebase origin/%s~1", portalBranch))
			},
			Compensate: func(input string) (string, error) {
				return "", nil
			},
		},
		{
			Name: "git reset commits",
			Run: func() (string, error) {
				return shell.Execute("git reset HEAD^")
			},
			Compensate: func(input string) (string, error) {
				return "", nil
			},
		},
		{
			Name: "delete remote portal branch",
			Run: func() (string, error) {
				return shell.Execute(fmt.Sprintf("git push origin --delete %s --progress", portalBranch))
			},
			Compensate: func(input string) (string, error) {
				return "", nil
			},
		},
	}
}
