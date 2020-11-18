package main

import (
	"fmt"
	"strings"
)

func currentBranchRemotelyUntracked() bool {
	_, err := execute("git rev-parse --abbrev-ref --symbolic-full-name @{u}")
	return err != nil
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

func localBranchExists(branch string) bool {
	command := fmt.Sprintf("git branch --list %s", branch)
	localBranch, err := execute(command)

	if err != nil {
		commandFailure(command, err)
	}

	return len(localBranch) > 0
}

func remoteBranchExists(branch string) bool {
	command := fmt.Sprintf("git ls-remote --heads origin %s", branch)
	remoteBranch, err := execute(command)

	if err != nil {
		commandFailure(command, err)
	}

	return len(remoteBranch) > 0
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

func buildPatch(remoteTrackingBranch string) (string, error) {
	return execute(fmt.Sprintf("git format-patch %s --stdout", remoteTrackingBranch))
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

func addAll(files string) (string, error) {
	return execute(fmt.Sprintf("git add %s", files))
}

func commit(message string) (string, error) {
	return execute(fmt.Sprintf("git commit --allow-empty -m %s", message))
}

func fetch() (string, error) {
	return execute("git fetch")
}

func showFile(branch string, fileName string) (string, error) {
	return execute(fmt.Sprintf("git show origin/%s:%s", branch, fileName))
}

func rebase() (string, error) {
	return execute("git pull -r")
}

func checkoutFile(branch string, filename string) (string, error) {
	return execute(fmt.Sprintf("git checkout origin/%s -- %s", branch, filename))
}
