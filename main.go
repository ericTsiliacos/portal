package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ericTsiliacos/portal/logger"
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

const EMPTY_INDEX = "nothing to push!"
const PORTAL_CLOSED = "nothing to pull!"
const REMOTE_TRACKING_REQUIRED = "must be on a branch that is remotely tracked"

func LOCAL_BRANCH_EXISTS(branch string) string {
	return fmt.Sprintf("local branch %s already exists", branch)
}

func REMOTE_BRANCH_EXISTS(branch string) string {
	return fmt.Sprintf("remote branch %s already exists", branch)
}

func DIRTY_INDEX(branch string) string {
	return fmt.Sprintf("%s: git index dirty!", branch)
}

func BRANCH_MISMATCH(startingBranch string, workingBranch string) string {
	return fmt.Sprintf("Starting branch %s did not match target branch %s", startingBranch, workingBranch)
}

func main() {

	defer logger.CloseLogOutput()

	commando.
		SetExecutableName("portal").
		SetVersion(version).
		SetDescription("A commandline tool for moving work-in-progress to and from your pair using git")

	commando.
		Register("push").
		SetShortDescription("push work-in-progress to pair").
		SetDescription("This command pushes work-in-progress to a branch for your pair to pull.").
		AddFlag("verbose,v", "displays commands and outputs", commando.Bool, false).
		AddFlag("strategy,s", "strategy to use for branch name: git-duet, git-together", commando.String, "auto").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {

			logger.LogInfo.Println(fmt.Sprintf("Version: %s", version))

			verbose, _ := flags["verbose"].GetBool()
			strategy, _ := flags["strategy"].GetString()

			validate(dirtyIndex() || unpublishedWork(), EMPTY_INDEX)
			validate(currentBranchRemotelyTracked(), REMOTE_TRACKING_REQUIRED)

			branchStrategy, err := branchNameStrategy(strategy)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			portalBranch := getPortalBranch(branchStrategy)

			validate(!localBranchExists(portalBranch), LOCAL_BRANCH_EXISTS(portalBranch))
			validate(!remoteBranchExists(portalBranch), REMOTE_BRANCH_EXISTS(portalBranch))

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
		AddFlag("strategy,s", "strategy to use for branch name: git-duet, git-together", commando.String, "auto").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {

			logger.LogInfo.Println(fmt.Sprintf("Version: %s", version))

			verbose, _ := flags["verbose"].GetBool()
			strategy, _ := flags["strategy"].GetString()

			validate(currentBranchRemotelyTracked(), REMOTE_TRACKING_REQUIRED)

			startingBranch := getCurrentBranch()

			validate(!dirtyIndex() && !unpublishedWork(), DIRTY_INDEX(startingBranch))

			branchStrategy, err := branchNameStrategy(strategy)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			portalBranch := getPortalBranch(branchStrategy)

			validate(remoteBranchExists(portalBranch), PORTAL_CLOSED)

			_, _ = fetch()

			metaFileContents, _ := showFile(portalBranch, "portal-meta.yml")
			config, _ := getConfiguration(metaFileContents)
			workingBranch := config.Meta.WorkingBranch
			pusherVersion := semver.Canonical(config.Meta.Version)
			sha := config.Meta.Sha
			currentVersion := semver.Canonical(version)

			if semver.Major(pusherVersion) != semver.Major(currentVersion) {
				fmt.Println("Pusher and Puller are using different versions of portal")
				fmt.Println("  1. Pusher run portal pull to retrieve changes.")
				fmt.Println("  2. Both pairs update to latest version of portal.")
				fmt.Println("\nThen try again...")
				os.Exit(1)
			}

			validate(workingBranch == startingBranch, BRANCH_MISMATCH(startingBranch, workingBranch))

			commands := []string{
				fmt.Sprintf("git rebase origin/%s", workingBranch),
				fmt.Sprintf("git reset --hard %s", sha),
				fmt.Sprintf("git rebase origin/%s~1", portalBranch),
				"git reset HEAD^",
				fmt.Sprintf("git push origin --delete %s --progress", portalBranch),
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

func getConfiguration(yamlContent string) (*config, error) {
	c := &config{}
	err := yaml.Unmarshal([]byte(yamlContent), c)
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

func branchNameStrategy(strategy string) ([]string, error) {
	strategies := map[string]interface{}{
		"git-duet":     gitDuet,
		"git-together": gitTogether,
	}

	if strategy != "auto" {
		fn, ok := strategies[strategy]
		if !ok {
			return nil, errors.New("unknown strategy")
		} else {
			return fn.(func() []string)(), nil
		}
	}

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

func validate(valid bool, message string) {
	if valid == false {
		fmt.Println(message)
		os.Exit(1)
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
