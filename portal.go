package main

import (
	"fmt"
	"github.com/progrium/go-shell"
	"github.com/thatisuday/commando"
	"sort"
	"strings"
)

var (
	sh = shell.Run
)

func main() {
	commando.
		SetExecutableName("portal").
		SetVersion("1.0.0").
		SetDescription("A commandline tool for moving work in progress to another machine using git")

	commando.
		Register("push").
		SetShortDescription("push work in progress").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			fmt.Println("Coming your way...")

			branch := branchName()

			sh(fmt.Sprintf("git checkout -b %s", branch))
			sh("git add .")
			sh("git commit -m \"WIP\"")
			sh("git push origin HEAD")
			sh("git checkout -")
			sh(fmt.Sprintf("git branch -D %s", branch))
		})

	commando.
		Register("pull").
		SetShortDescription("pull work in progress").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			fmt.Println("Coming your way...")

			branch := branchName()

			sh(fmt.Sprintf("git checkout -b %s origin/%s", branch, branch))
			sh("git reset HEAD^")
			sh("git checkout -")
			sh(fmt.Sprintf("git branch -D %s", branch))
			sh(fmt.Sprintf("git push origin --delete %s", branch))
		})

	commando.Parse(nil)
}

func branchName() string {
	author := sh("git config --get duet.env.git-author-initials")
	coauthor := sh("git config --get duet.env.git-committer-initials")

	authors := []string{author.Stdout.String(), coauthor.Stdout.String()}
	sort.Strings(authors)
	authors = Map(authors, func(s string) string {
		return strings.TrimSuffix(s, "\n")
	})
	authors = append([]string{"portal"}, authors...)
	branch := strings.Join(authors, "-")
	return branch
}

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}
