package main

import (
	"fmt"
	"os"
	"time"

	"github.com/briandowns/spinner"
	"github.com/ericTsiliacos/portal/internal/constants"
	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/logger"
	"github.com/ericTsiliacos/portal/internal/portal"
	"github.com/ericTsiliacos/portal/internal/shell"
	"github.com/thatisuday/commando"
	"golang.org/x/mod/semver"
)

var version string

func main() {

	defer logger.CloseLogOutput()

	commando.
		SetExecutableName("portal").
		SetVersion(version).
		SetDescription("A commandline tool for moving work between pairs using git")

	commando.
		Register("push").
		SetDescription("Push changes to a portal branch").
		AddFlag("verbose,v", "verbose output", commando.Bool, false).
		AddFlag("strategy,s", "git-duet, git-together", commando.String, "auto").
		AddFlag("patch,p", "create and stash patch", commando.Bool, false).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {

			logger.LogInfo.Println(fmt.Sprintf("Version: %s", version))

			verbose, _ := flags["verbose"].GetBool()
			strategy, _ := flags["strategy"].GetString()
			patch, _ := flags["patch"].GetBool()

			portalBranch, err := portal.BranchNameStrategy(strategy)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			validate(git.DirtyIndex() || git.UnpublishedWork(), constants.EMPTY_INDEX)
			validate(git.CurrentBranchRemotelyTracked(), constants.REMOTE_TRACKING_REQUIRED)
			validate(!git.LocalBranchExists(portalBranch), constants.LOCAL_BRANCH_EXISTS(portalBranch))
			validate(!git.RemoteBranchExists(portalBranch), constants.REMOTE_BRANCH_EXISTS(portalBranch))

			currentBranch := git.GetCurrentBranch()
			remoteTrackingBranch := git.GetRemoteTrackingBranch()
			sha := git.GetBoundarySha(remoteTrackingBranch, currentBranch)

			_, _ = git.Add(".")
			_, _ = git.Commit("portal-wip")

			savePatch(patch, remoteTrackingBranch, func() {
				portal.WritePortalMetaData(currentBranch, sha, version)
				_, _ = git.Add("portal-meta.yml")
				_, _ = git.Commit("portal-meta")
			})

			commands := []string{
				fmt.Sprintf("git checkout -b %s --progress", portalBranch),
				fmt.Sprintf("git push origin %s --progress", portalBranch),
				fmt.Sprintf("git checkout %s --progress", currentBranch),
				fmt.Sprintf("git reset --hard %s", remoteTrackingBranch),
				fmt.Sprintf("git branch -D %s", portalBranch),
			}

			runner(commands, verbose)

			fmt.Println("✨ Sent!")
		})

	commando.
		Register("pull").
		SetDescription("Pull changes from portal branch").
		AddFlag("verbose,v", "verbose output", commando.Bool, false).
		AddFlag("strategy,s", "git-duet, git-together", commando.String, "auto").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {

			logger.LogInfo.Println(fmt.Sprintf("Version: %s", version))

			verbose, _ := flags["verbose"].GetBool()
			strategy, _ := flags["strategy"].GetString()

			portalBranch, err := portal.BranchNameStrategy(strategy)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			validate(git.RemoteBranchExists(portalBranch), constants.PORTAL_CLOSED)

			_, _ = git.Fetch()

			metaFileContents, _ := git.ShowFile(portalBranch, "portal-meta.yml")
			config, _ := portal.GetConfiguration(metaFileContents)
			workingBranch := config.Meta.WorkingBranch
			pusherVersion := semver.Canonical(config.Meta.Version)
			sha := config.Meta.Sha
			pullerVersion := semver.Canonical(version)

			validate(semver.Major(pusherVersion) == semver.Major(pullerVersion), constants.DIFFERENT_VERSIONS)
			validate(git.CurrentBranchRemotelyTracked(), constants.REMOTE_TRACKING_REQUIRED)
			startingBranch := git.GetCurrentBranch()
			validate(workingBranch == startingBranch, constants.BRANCH_MISMATCH(startingBranch, workingBranch))
			validate(!git.DirtyIndex() && !git.UnpublishedWork(), constants.DIRTY_INDEX(startingBranch))

			commands := []string{
				fmt.Sprintf("git rebase origin/%s", workingBranch),
				fmt.Sprintf("git reset --hard %s", sha),
				fmt.Sprintf("git rebase origin/%s~1", portalBranch),
				"git reset HEAD^",
				fmt.Sprintf("git push origin --delete %s --progress", portalBranch),
			}

			runner(commands, verbose)

			fmt.Println("✨ Got it!")
		})

	commando.Parse(nil)
}

func runner(commands []string, verbose bool) {
	if verbose {
		run(commands, verbose)
	} else {
		style(func() {
			run(commands, verbose)
		})
	}
}

func run(commands []string, verbose bool) {
	for _, command := range commands {
		if verbose {
			fmt.Println(command)
		}

		output := shell.Check(shell.Execute(command))

		if verbose {
			fmt.Println(output)
		}
	}
}

func style(fn func()) {
	s := spinner.New(spinner.CharSets[23], 100*time.Millisecond)
	s.Suffix = " Coming your way..."
	s.Start()

	fn()

	s.Stop()
}

func savePatch(save bool, remoteTrackingBranch string, fn func()) {
	if save {
		now := time.Now().Format(time.RFC3339)
		portal.Patch(remoteTrackingBranch, now)

		fn()

		shell.Execute(fmt.Sprintf("git stash save \"portal-patch-%s\" --include-untracked", now))
	} else {
		fn()
	}
}

func validate(valid bool, message string) {
	if !valid {
		fmt.Println(message)
		os.Exit(1)
	}
}
