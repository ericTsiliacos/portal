package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
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

const (
	exitCodeInterrupt = 2
)

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

			logger.LogInfo.Println(fmt.Sprintf("Portal: %s", version))

			commitMessage := os.Getenv("PORTAL_COMMIT_MESSAGE")

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

			ctx, cancel, signalChan := cancelContext()
			defer stop(cancel, signalChan)
			go handleCancel(ctx, cancel, signalChan)

			pushSteps, err := portal.PushSagaSteps(ctx, portalBranch, version, verbose, commitMessage)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			errors := stylized(verbose, func() []string {
				saga := saga.New(pushSteps)
				return saga.Run()
			})

			if errors != nil {
				for _, error := range errors {
					fmt.Println(error)
				}
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

			logger.LogInfo.Println(fmt.Sprintf("Portal: %s", version))

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
			startingBranch, err := git.GetCurrentBranch()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			validate(workingBranch == startingBranch, constants.BRANCH_MISMATCH(startingBranch, workingBranch))
			validate(!git.DirtyIndex() && !git.UnpublishedWork(), constants.DIRTY_INDEX(startingBranch))

			ctx, cancel, signalChan := cancelContext()
			defer stop(cancel, signalChan)
			go handleCancel(ctx, cancel, signalChan)

			pullSteps, err := portal.PullSagaSteps(ctx, startingBranch, portalBranch, sha, verbose)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			errors := stylized(verbose, func() []string {
				saga := saga.New(pullSteps)
				return saga.Run()
			})

			if errors != nil {
				for _, error := range errors {
					fmt.Println(error)
				}
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

func cancelContext() (context.Context, context.CancelFunc, chan os.Signal) {
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	return ctx, cancel, signalChan
}

func stop(cancel context.CancelFunc, signalChan chan os.Signal) {
	signal.Stop(signalChan)
	cancel()
}

func handleCancel(ctx context.Context, cancel context.CancelFunc, signalChan chan os.Signal) {
	select {
	case <-signalChan:
		cancel()
	case <-ctx.Done():
	}
	<-signalChan
	os.Exit(exitCodeInterrupt)
}
