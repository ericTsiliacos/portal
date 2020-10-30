package main

import "C"
import (
	"errors"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/thatisuday/commando"
	"log"
	"os"
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

			branchStrategy, err := branchNameStrategy()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			branch := branchName(branchStrategy)

			checkRemoteBranchExistence(branch)

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

			checkForDirtyIndex()

			branchStrategy, err := branchNameStrategy()
			if err != nil {
				log.Fatalf("Error: %v", err)
			}

			branch := branchName(branchStrategy)

			commands := []string{
				"git fetch",
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

func branchNameStrategy() ([]string, error) {
	gitDuetAuthors := gitDuet()
	if len(gitDuetAuthors) > 1 {
		return gitDuetAuthors, nil
	}

	gitTogetherAuthors := gitTogether()
	if len(gitTogetherAuthors) > 1 {
		return gitTogetherAuthors, nil
	}

	return nil, errors.New("branch naming strategy not found")
}

func checkRemoteBranchExistence(branch string) {
	command := fmt.Sprintf("git ls-remote --heads origin %s", branch)
	remoteBranch, err := execute(command)

	if err != nil {
		commandFailure(command, err)
	}

	if len(remoteBranch) > 0 {
		fmt.Println(fmt.Sprintf("remote branch %s already exists", branch))
		os.Exit(1)
	}
}

func checkForDirtyIndex() {
	command := "git status --porcelain=v1"
	index, err := execute(command)

	if err != nil {
		commandFailure(command, err)
	}

	indexCount := strings.Count(index, "\n")
	if indexCount > 0 {
		fmt.Println("git index dirty!")
		os.Exit(1)
	}
}

func runner(commands []string, dryRun bool, verbose bool, completionMessage string) {
	if dryRun == true {
		runDry(commands)
	} else {
		if verbose == true {
			run(commands, verbose)
		} else {
			style(func() {
				run(commands, verbose)
			})
		}

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
		output, err := execute(command)

		if err != nil {
			commandFailure(command, err)
		}

		if verbose == true {
			fmt.Println(output)
		}
	}
}

func runDry(commands []string) {
	for _, command := range commands {
		fmt.Println(command)
	}
}

func branchName(authors []string) string {
	sort.Strings(authors)
	authors = Map(authors, func(s string) string {
		return strings.TrimSuffix(s, "\n")
	})
	branch := strings.Join(authors, "-")
	return portal(branch)
}

func portal(branchName string) string {
	return "portal-" + branchName
}

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func commandFailure(command string, err error) {
	fmt.Println(command)
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
