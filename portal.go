package main

import (
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/thatisuday/commando"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

func main() {
	commando.
		SetExecutableName("portal").
		SetVersion("1.0.0").
		SetDescription("A commandline tool for moving work-in-progress to and from your pair using git")

	commando.
		Register("push").
		SetShortDescription("push work-in-progress to pair").
		SetDescription("This command pushes work-in-progress to a branch for your pair to pull.").
		AddFlag("dry-run,n", "list of commands to run side-effects free", commando.Bool, false).
		AddFlag("verbose,v", "displays commands and outputs", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			dryRun, _ := flags["dry-run"].GetBool()
			verbose, _ := flags["verbose"].GetBool()

			branch := branchName()

			commands := []string{
				fmt.Sprintf("git checkout -b %s", branch),
				"git add .",
				"git commit -m \"WIP\"",
				"git push origin HEAD",
				"git checkout -",
				fmt.Sprintf("git branch -D %s", branch),
			}

			runner(commands, dryRun, verbose, "✨ Sent!")
		})

	commando.
		Register("pull").
		SetShortDescription("pull work-in-progress from pair").
		SetDescription("This command pulls work-in-progress from your pair.").
		AddFlag("dry-run,n", "list of commands to run side-effects free", commando.Bool, false).
		AddFlag("verbose,v", "displays commands and outputs", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			dryRun, _ := flags["dry-run"].GetBool()
			verbose, _ := flags["verbose"].GetBool()

			branch := branchName()

			commands := []string{
				fmt.Sprintf("git checkout -b %s origin/%s", branch, branch),
				"git reset HEAD^",
				"git checkout -",
				fmt.Sprintf("git branch -D %s", branch),
				fmt.Sprintf("git push origin --delete %s", branch),
			}

			runner(commands, dryRun, verbose, "✨ Got it!")
		})

	commando.Parse(nil)
}

func runner(commands []string, dryRun bool, verbose bool, completionMessage string) {
	if dryRun == true {
		runDry(commands)
	} else {
		style(func() {
			run(commands, verbose)
		})

		fmt.Println(completionMessage)
	}
}

type terminal func()

func style(fn terminal) {
	s := spinner.New(spinner.CharSets[23], 100*time.Millisecond)
	s.Suffix = " Coming your way..."
	s.Start()

	fn()

	s.Stop()
}

func run(commands []string, verbose bool) {
	for _, command := range commands {
		if verbose == true {
			fmt.Println(command)
		}
		output := execute(command)

		if verbose == true {
			fmt.Println(output)
		}
	}
}

func execute(command string) string {
	s := strings.Split(command, " ")
	cmd, args := s[0], s[1:]
	cmdOut, err := exec.Command(cmd, args...).Output()

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return string(cmdOut)
}

func runDry(commands []string) {
	for _, command := range commands {
		fmt.Println(command)
	}
}

func branchName() string {
	author := execute("git config --get duet.env.git-author-initials")
	coauthor := execute("git config --get duet.env.git-committer-initials")

	authors := []string{author, coauthor}
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
