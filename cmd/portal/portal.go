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
	"github.com/ericTsiliacos/portal/internal/saga"
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
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {

			logger.LogInfo.Println(fmt.Sprintf("Version: %s", version))
			logger.LogInfo.Println(git.Version())

			verbose, _ := flags["verbose"].GetBool()
			strategy, _ := flags["strategy"].GetString()

			portalBranch, err := portal.BranchNameStrategy(strategy)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			validate(git.CurrentWorkingDirectoryGitRoot(), constants.GIT_ROOT)
			validate(git.DirtyIndex() || git.UnpublishedWork(), constants.EMPTY_INDEX)
			validate(git.CurrentBranchRemotelyTracked(), constants.REMOTE_TRACKING_REQUIRED)
			validate(!git.LocalBranchExists(portalBranch), constants.LOCAL_BRANCH_EXISTS(portalBranch))
			validate(!git.RemoteBranchExists(portalBranch), constants.REMOTE_BRANCH_EXISTS(portalBranch))

			now := time.Now().Format(time.RFC3339)

			pushSteps := portal.PushSagaSteps(portalBranch, now, version)

			errors := stylized(verbose, func() []string {
				saga := saga.New(pushSteps)
				return saga.Run()
			})

			if errors != nil {
				fmt.Println(errors)
			} else {
				fmt.Println("✨ Sent!")
			}
		})

	commando.
		Register("pull").
		SetDescription("Pull changes from portal branch").
		AddFlag("verbose,v", "verbose output", commando.Bool, false).
		AddFlag("strategy,s", "git-duet, git-together", commando.String, "auto").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {

			logger.LogInfo.Println(fmt.Sprintf("Version: %s", version))
			logger.LogInfo.Println(git.Version())

			verbose, _ := flags["verbose"].GetBool()
			strategy, _ := flags["strategy"].GetString()

			portalBranch, err := portal.BranchNameStrategy(strategy)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			validate(git.CurrentWorkingDirectoryGitRoot(), constants.GIT_ROOT)
			validate(git.RemoteBranchExists(portalBranch), constants.PORTAL_CLOSED)

			_, _ = git.Fetch()

			metaFileContents, _ := git.ShowCommitMessage(portalBranch)
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

			pullSteps := portal.PullSagaSteps(startingBranch, portalBranch, sha)

			errors := stylized(verbose, func() []string {
				saga := saga.New(pullSteps)
				return saga.Run()
			})

			if errors != nil {
				fmt.Println(errors)
			} else {
				fmt.Println("✨ Got it!")
			}
		})

	commando.Parse(nil)
}

func stylized(verbose bool, fn func() []string) []string {
	if verbose {
		return fn()
	} else {
		s := spinner.New(spinner.CharSets[23], 100*time.Millisecond)
		s.Suffix = " Coming your way..."
		s.Start()

		err := fn()

		s.Stop()

		return err
	}
}

func validate(valid bool, message string) {
	if !valid {
		fmt.Println(message)
		os.Exit(1)
	}
}
