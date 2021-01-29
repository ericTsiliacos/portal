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
			logger.LogInfo.Println(git.Version())

			// verbose, _ := flags["verbose"].GetBool()
			strategy, _ := flags["strategy"].GetString()
			// patch, _ := flags["patch"].GetBool()

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

			originalBranch := git.GetCurrentBranch()
			remoteTrackingBranch := git.GetRemoteTrackingBranch()
			sha := git.GetBoundarySha(remoteTrackingBranch, originalBranch)
			now := time.Now().Format(time.RFC3339)

			steps := []saga.Step{
				{
					Name:       "git add .",
					Run:        func() (string, error) { return git.Add(".") },
					Compensate: func(input string) (string, error) { return git.Reset() },
				},
				{
					Name:       "git commit -m 'portal-wip'",
					Run:        func() (string, error) { return git.Commit("portal-wip") },
					Compensate: func(input string) (string, error) { return git.UndoCommit() },
				},
				{
					Name: "back up work in progress",
					Run: func() (string, error) {
						return portal.Patch(remoteTrackingBranch, now)
					},
					Compensate: func(fileName string) (string, error) {
						return shell.Execute(fmt.Sprintf("rm %s", fileName))
					},
				},
				{
					Name: "create portal metadata file",
					Run: func() (string, error) {
						return portal.WritePortalMetadata("portal-meta.yml", originalBranch, sha, version)
					},
					Compensate: func(fileName string) (string, error) {
						return shell.Execute(fmt.Sprintf("rm %s", fileName))
					},
				},
				{
					Name:       "git add portal-meta.yml",
					Run:        func() (string, error) { return git.Add("portal-meta.yml") },
					Compensate: func(input string) (string, error) { return git.Reset() },
				},
				{
					Name:       "git commit -m 'portal-meta'",
					Run:        func() (string, error) { return git.Commit("portal-meta") },
					Compensate: func(input string) (string, error) { return git.UndoCommit() },
				},
				{
					Name: "git stash save backup patch",
					Run: func() (string, error) {
						return shell.Execute(fmt.Sprintf("git stash push -m \"portal-patch-%s\" --include-untracked", now))
					},
					Compensate: func(input string) (string, error) {
						return shell.Execute("git stash pop")
					},
				},
				{
					Name: "git checkout portal branch",
					Run: func() (string, error) {
						_, err := shell.Execute(fmt.Sprintf("git checkout -b %s --progress", portalBranch))
						return portalBranch, err
					},
					Compensate: func(portalBranch string) (string, error) {
						shell.Execute(fmt.Sprintf("git checkout %s --progress", originalBranch))
						return shell.Execute(fmt.Sprintf("git branch -D %s", portalBranch))
					},
				},
				{
					Name: "git push portal branch",
					Run: func() (string, error) {
						return shell.Execute(fmt.Sprintf("git push origin %s --progress", portalBranch))
					},
					Compensate: func(originalBranch string) (string, error) {
						return shell.Execute(fmt.Sprintf("git push origin --delete %s --progress", portalBranch))
					},
					Retries: 1,
				},
				{
					Name: "git checkout to original branch",
					Run: func() (string, error) {
						return shell.Execute(fmt.Sprintf("git checkout %s --progress", originalBranch))
					},
				},
				{
					Name: "delete local portal branch",
					Run: func() (string, error) {
						return shell.Execute(fmt.Sprintf("git branch -D %s", portalBranch))
					},
				},
				{
					Name: "clear git workspace",
					Run: func() (string, error) {
						return shell.Execute(fmt.Sprintf("git reset --hard %s", remoteTrackingBranch))
					},
				},
			}
			s := spinner.New(spinner.CharSets[23], 100*time.Millisecond)
			s.Suffix = " Coming your way..."
			s.Start()

			saga := saga.Saga{Steps: steps}
			err = saga.Run()

			s.Stop()

			// savePatch(patch, remoteTrackingBranch, func() {
			// 	portal.WritePortalMetadata("portal-meta.yml", originalBranch, sha, version)
			// 	_, _ = git.Add("portal-meta.yml")
			// 	_, _ = git.Commit("portal-meta")
			// })

			// commands := []string{
			// 	fmt.Sprintf("git checkout -b %s --progress", portalBranch),
			// 	fmt.Sprintf("git push origin %s --progress", portalBranch),
			// 	fmt.Sprintf("git checkout %s --progress", originalBranch),
			// 	fmt.Sprintf("git reset --hard %s", remoteTrackingBranch),
			// 	fmt.Sprintf("git branch -D %s", portalBranch),
			// }

			// runner(commands, verbose)

			if err != nil {
				fmt.Println(err)
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
