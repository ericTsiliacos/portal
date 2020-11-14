package main

import (
	"errors"
	"fmt"
	"github.com/briandowns/spinner"
	"github.com/thatisuday/commando"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

var version string

type config struct {
	Portal struct {
		Version       string `yaml:"version"`
		WorkingBranch string `yaml:"workingBranch"`
		Sha           string `yaml:"sha"`
	} `yaml:"Portal"`
}

func main() {
	commando.
		SetExecutableName("portal").
		SetVersion(version).
		SetDescription("A commandline tool for moving work-in-progress to and from your pair using git")

	commando.
		Register("push").
		SetShortDescription("push work-in-progress to pair").
		SetDescription("This command pushes work-in-progress to a branch for your pair to pull.").
		AddFlag("verbose,v", "displays commands and outputs", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			verbose, _ := flags["verbose"].GetBool()

			if !dirtyIndex() && !unpublishedWork() {
				fmt.Println("nothing to push!")
				os.Exit(1)
			}

			branchStrategy, err := branchNameStrategy()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			portalBranch := getPortalBranch(branchStrategy)

			err = checkCurrentBranchRemoteTracking()
			if err != nil {
				fmt.Println("Only branches with remote tracking are pushable")
				os.Exit(1)
			}

			checkLocalBranchNonExistence(portalBranch)
			checkRemoteBranchNonExistence(portalBranch)

			currentBranch := getCurrentBranch()

			remoteTrackingBranch := getRemoteTrackingBranch()

			sha := getBoundarySha(remoteTrackingBranch, currentBranch)

			_, _ = execute("git add .")
			_, _ = execute("git commit --allow-empty -m portal-wip")

			savePatch(remoteTrackingBranch)

			writePortalMetaData(currentBranch, sha, version)
			_, _ = execute("git add portal-meta.yml")
			_, _ = execute("git commit --allow-empty -m portal-meta")

			commands := []string{
				"git stash save \"portal-save-patch\" --include-untracked",
				fmt.Sprintf("git checkout -b %s --progress", portalBranch),
				fmt.Sprintf("git push origin %s --progress", portalBranch),
				fmt.Sprintf("git checkout %s --progress", currentBranch),
				fmt.Sprintf("git branch -D %s", portalBranch),
				fmt.Sprintf("git reset --hard %s", remoteTrackingBranch),
			}

			runner(commands, verbose, "✨ Sent!")
		})

	commando.
		Register("pull").
		SetShortDescription("pull work-in-progress from pair").
		SetDescription("This command pulls work-in-progress from your pair.").
		AddFlag("verbose,v", "displays commands and outputs", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			verbose, _ := flags["verbose"].GetBool()

			err := checkCurrentBranchRemoteTracking()
			if err != nil {
				fmt.Println("Must be on a branch that is remotely tracked.")
				os.Exit(1)
			}

			startingBranch := getCurrentBranch()

			if dirtyIndex() || unpublishedWork() {
				fmt.Println(fmt.Sprintf("%s: git index dirty!", startingBranch))
				os.Exit(1)
			}

			branchStrategy, err := branchNameStrategy()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			portalBranch := getPortalBranch(branchStrategy)

			checkRemoteBranchExistence(portalBranch)

			_, err = execute("git fetch")
			if err != nil {
				fmt.Println("failed fetch")
				os.Exit(1)
			}

			_, err = execute("git pull -r")
			if err != nil {
				fmt.Println("failed to pull rebase")
				os.Exit(1)
			}

			_, err = execute(fmt.Sprintf("git checkout origin/%s -- portal-meta.yml", portalBranch))
			if err != nil {
				fmt.Println("failed to checkout meta file")
				os.Exit(1)
			}

			config, _ := getConfiguration()
			workingBranch := config.Portal.WorkingBranch
			pusherVersion := config.Portal.Version
			sha := config.Portal.Sha

			_, err = execute("rm portal-meta.yml")
			if err != nil {
				fmt.Println("failed to delete meta file")
				os.Exit(1)
			}

			if pusherVersion != version {
				fmt.Println("Pusher and Puller are using different versions of portal")
				fmt.Println("  1. Pusher run portal pull to retrieve changes.")
				fmt.Println("  2. Both pairs update to latest version of portal.")
				fmt.Println("\nThen try again...")
				os.Exit(1)
			}

			if workingBranch != startingBranch {
				fmt.Println(fmt.Sprintf("Starting branch %s did not match target branch %s", startingBranch, workingBranch))
				os.Exit(1)
			}

			execute(fmt.Sprintf("git reset --hard %s", sha))
			execute(fmt.Sprintf("git rebase origin/%s", portalBranch))
			execute("git reset HEAD^^")
			execute("rm portal-meta.yml")
			execute(fmt.Sprintf("git push origin --delete %s", portalBranch))

			commands := []string{
				"echo done",
			}

			runner(commands, verbose, "✨ Got it!")
		})

	commando.Parse(nil)
}

func parseRefBoundary(revisionBoundaries string) string {
	boundaries := strings.FieldsFunc(revisionBoundaries, func(c rune) bool {
		return c == '\n'
	})

	return trimFirstRune(boundaries[len(boundaries)-1])
}

func getConfiguration() (*config, error) {
	yamlFile, err := ioutil.ReadFile("portal-meta.yml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	c := &config{}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c, err
}

func writePortalMetaData(branch string, sha string, version string) {
	f, err := os.Create("portal-meta.yml")
	if err != nil {
		fmt.Println(err)
		return
	}

	c := config{}
	c.Portal.WorkingBranch = branch
	c.Portal.Sha = sha
	c.Portal.Version = version

	d, err := yaml.Marshal(&c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = f.WriteString(string(d))
	if err != nil {
		fmt.Println(err)
		_ = f.Close()
		return
	}
}

func branchNameStrategy() ([]string, error) {
	pairs := [][]string{
		gitDuet(),
		gitTogether(),
	}

	if all(pairs, empty) {
		return nil, errors.New("no branch naming strategy found")
	}

	if many(pairs, nonEmpty) {
		return nil, errors.New("multiple branch naming strategies found")
	}

	return findFirst(pairs, nonEmpty), nil
}

func runner(commands []string, verbose bool, completionMessage string) {
	if verbose == true {
		run(commands, verbose)
	} else {
		style(func() {
			run(commands, verbose)
		})
	}

	fmt.Println(completionMessage)
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

func getPortalBranch(authors []string) string {
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

func all(xxs [][]string, f func(xs []string) bool) bool {
	for _, xs := range xxs {
		if f(xs) == false {
			return false
		}
	}
	return true
}

func many(xxs [][]string, f func(xs []string) bool) bool {
	seen := false
	for _, xs := range xxs {
		if seen && f(xs) {
			return true
		} else if f(xs) {
			seen = true
		} else {
			continue
		}
	}
	return false
}

func findFirst(xxs [][]string, f func(xs []string) bool) []string {
	for i, xs := range xxs {
		if f(xs) == true {
			return xxs[i]
		}
	}

	return nil
}

func empty(xs []string) bool {
	return len(xs) == 0
}

func nonEmpty(xs []string) bool {
	return len(xs) != 0
}

func execute(command string) (string, error) {
	s := strings.Split(command, " ")
	cmd, args := s[0], s[1:]
	cmdOut, err := exec.Command(cmd, args...).CombinedOutput()
	return string(cmdOut), err
}

func commandFailure(command string, err error) {
	fmt.Println(command)
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func trimFirstRune(s string) string {
	_, i := utf8.DecodeRuneInString(s)
	return s[i:]
}

func savePatch(remoteTrackingBranch string) {
	patch, _ := execute(fmt.Sprintf("git format-patch %s --stdout", remoteTrackingBranch))
	f, _ := os.Create("portal.patch")
	_, _ = f.WriteString(patch)
	_ = f.Close()
}

func getBoundarySha(remoteTrackingBranch string, currentBranch string) string {
	revisionBoundaries, _ := execute(fmt.Sprintf("git rev-list --boundary %s..%s", remoteTrackingBranch, currentBranch))
	if len(revisionBoundaries) > 0 {
		return parseRefBoundary(revisionBoundaries)
	} else {
		currentRev, _ := execute(fmt.Sprintf("git rev-parse %s", currentBranch))
		cleanCurrentRev := strings.TrimSuffix(currentRev, "\n")
		return cleanCurrentRev
	}
}

func checkCurrentBranchRemoteTracking() error {
	_, err := execute("git rev-parse --abbrev-ref --symbolic-full-name @{u}")
	return err
}

func getCurrentBranch() string {
	currentBranch, _ := execute("git rev-parse --abbrev-ref HEAD")
	cleanCurrentBranch := strings.TrimSuffix(currentBranch, "\n")
	return cleanCurrentBranch
}

func getRemoteTrackingBranch() string {
	remoteTrackingBranch, _ := execute("git rev-parse --abbrev-ref --symbolic-full-name @{u}")
	cleanRemoteTrackingBranch := strings.TrimSuffix(remoteTrackingBranch, "\n")
	return cleanRemoteTrackingBranch
}

func gitDuet() []string {
	author, authorErr := execute("git config --get duet.env.git-author-initials")
	coauthor, coauthorErr := execute("git config --get duet.env.git-committer-initials")

	if authorErr != nil && coauthorErr != nil {
		return []string{}
	}

	return []string{author, coauthor}
}

func gitTogether() []string {
	activeAuthors, err := execute("git config --get git-together.active")

	if err != nil {
		return []string{}
	}

	return strings.Split(activeAuthors, "+")
}

func checkLocalBranchNonExistence(branch string) {
	command := fmt.Sprintf("git branch --list %s", branch)
	localBranch, err := execute(command)

	if err != nil {
		commandFailure(command, err)
	}

	if len(localBranch) > 0 {
		fmt.Println(fmt.Sprintf("local branch %s already exists", branch))
		os.Exit(1)
	}
}

func checkRemoteBranchNonExistence(branch string) {
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

func checkRemoteBranchExistence(branch string) {
	command := fmt.Sprintf("git ls-remote --heads origin %s", branch)
	remoteBranch, err := execute(command)

	if err != nil {
		commandFailure(command, err)
	}

	if len(remoteBranch) == 0 {
		fmt.Println("nothing to pull!")
		os.Exit(1)
	}
}

func dirtyIndex() bool {
	command := "git status --porcelain=v1"
	index, err := execute(command)

	if err != nil {
		commandFailure(command, err)
	}

	indexCount := strings.Count(index, "\n")
	return indexCount > 0
}

func unpublishedWork() bool {
	command := "git status -sb"
	output, err := execute(command)

	if err != nil {
		commandFailure(command, err)
	}

	return strings.Contains(output, "ahead")
}
