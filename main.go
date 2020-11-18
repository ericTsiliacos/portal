package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/thatisuday/commando"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
)

var version string

type config struct {
	Meta struct {
		Version       string `yaml:"version"`
		WorkingBranch string `yaml:"workingBranch"`
		Sha           string `yaml:"sha"`
	} `yaml:"Meta"`
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

			if currentBranchRemotelyUntracked() {
				fmt.Println("Only branches with remote tracking are pushable")
				os.Exit(1)
			}

			branchStrategy, err := branchNameStrategy()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			portalBranch := getPortalBranch(branchStrategy)

			if localBranchExists(portalBranch) {
				fmt.Println(fmt.Sprintf("local branch %s already exists", portalBranch))
				os.Exit(1)
			}

			if remoteBranchExists(portalBranch) {
				fmt.Println(fmt.Sprintf("remote branch %s already exists", portalBranch))
				os.Exit(1)
			}

			currentBranch := getCurrentBranch()

			remoteTrackingBranch := getRemoteTrackingBranch()

			sha := getBoundarySha(remoteTrackingBranch, currentBranch)

			_, _ = addAll(".")
			_, _ = commit("portal-wip")

			now := time.Now().Format(time.RFC3339)
			savePatch(remoteTrackingBranch, now)

			writePortalMetaData(currentBranch, sha, version)
			_, _ = addAll("portal-meta.yml")
			_, _ = commit("portal-meta")

			commands := []string{
				fmt.Sprintf("git stash save \"portal-patch-%s\" --include-untracked", now),
				fmt.Sprintf("git checkout -b %s --progress", portalBranch),
				fmt.Sprintf("git push origin %s --progress", portalBranch),
				fmt.Sprintf("git checkout %s --progress", currentBranch),
				fmt.Sprintf("git reset --hard %s", remoteTrackingBranch),
				fmt.Sprintf("git branch -D %s", portalBranch),
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

			if currentBranchRemotelyUntracked() {
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

			if !remoteBranchExists(portalBranch) {
				fmt.Println("nothing to pull!")
				os.Exit(1)
			}

			_, _ = fetch()
			_, _ = rebase()

			metaFilename := "portal-meta.yml"
			_, _ = checkoutFile(portalBranch, metaFilename)

			config, _ := getConfiguration()
			workingBranch := config.Meta.WorkingBranch
			pusherVersion := semver.Canonical(config.Meta.Version)
			sha := config.Meta.Sha
			currentVersion := semver.Canonical(version)

			_, _ = removeFile(metaFilename)

			if semver.Major(pusherVersion) != semver.Major(currentVersion) {
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

			commands := []string{
				fmt.Sprintf("git reset --hard %s", sha),
				fmt.Sprintf("git rebase origin/%s~1", portalBranch),
				"git reset HEAD^",
				fmt.Sprintf("git push origin --delete %s", portalBranch),
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
	c.Meta.WorkingBranch = branch
	c.Meta.Sha = sha
	c.Meta.Version = version

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

func savePatch(remoteTrackingBranch string, dateTime string) {
	patch, _ := buildPatch(remoteTrackingBranch)
	f, _ := os.Create(buildPatchFileName(dateTime))
	_, _ = f.WriteString(patch)
	_ = f.Close()
}

func buildPatchFileName(dateTime string) string {
	return fmt.Sprintf("portal-%s.patch", dateTime)
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

func runDry(commands []string) {
	for _, command := range commands {
		fmt.Println(command)
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

func removeFile(filename string) (string, error) {
	return execute(fmt.Sprintf("rm %s", filename))
}
