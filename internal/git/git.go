package git

import (
	"fmt"
	"strings"

	"github.com/ericTsiliacos/portal/internal/char"
	"github.com/ericTsiliacos/portal/internal/shell"
)

func CurrentBranchRemotelyTracked() bool {
	remoteBranch, err := shell.Execute("git rev-parse --abbrev-ref --symbolic-full-name @{u}")
	return len(remoteBranch) > 0 && err == nil
}

func DirtyIndex() bool {
	index := shell.Check(shell.Execute("git status --porcelain=v1"))

	indexCount := strings.Count(index, "\n")
	return indexCount > 0
}

func IsGitProject() bool {
	output, err := shell.Execute("git rev-parse --is-inside-work-tree")

	if err != nil {
		return false
	}

	isGitProject := strings.TrimSuffix(output, "\n")

	return isGitProject == "true"
}

func UnpublishedWork() bool {
	output := shell.Check(shell.Execute("git status -sb"))

	return strings.Contains(output, "ahead")
}

func LocalBranchExists(branch string) bool {
	localBranch := shell.Check(shell.Execute(fmt.Sprintf("git branch --list %s", branch)))

	return len(localBranch) > 0
}

func RemoteBranchExists(branch string) bool {
	remoteBranch := shell.Check(shell.Execute(fmt.Sprintf("git ls-remote --heads origin %s", branch)))

	return len(remoteBranch) > 0
}

func GitDuet() []string {
	author, authorErr := shell.Execute("git config --get duet.env.git-author-initials")
	coauthor, coauthorErr := shell.Execute("git config --get duet.env.git-committer-initials")

	if authorErr != nil && coauthorErr != nil {
		return []string{}
	}

	return []string{author, coauthor}
}

func GitTogether() []string {
	activeAuthors, err := shell.Execute("git config --get git-together.active")

	if err != nil {
		return []string{}
	}

	return strings.Split(activeAuthors, "+")
}

func GetCurrentBranch() (string, error) {
	currentBranch, err := shell.Execute("git rev-parse --abbrev-ref HEAD")
	if err != nil {
		return "", err
	}
	cleanCurrentBranch := strings.TrimSuffix(currentBranch, "\n")
	return cleanCurrentBranch, nil
}

func GetRemoteTrackingBranch() (string, error) {
	remoteTrackingBranch, err := shell.Execute("git rev-parse --abbrev-ref --symbolic-full-name @{u}")
	if err != nil {
		return "", err
	}
	cleanRemoteTrackingBranch := strings.TrimSuffix(remoteTrackingBranch, "\n")
	return cleanRemoteTrackingBranch, nil
}

func GetBoundarySha(remoteTrackingBranch string, currentBranch string) (string, error) {
	revisionBoundaries, err := shell.Execute(fmt.Sprintf("git rev-list --boundary %s..%s", remoteTrackingBranch, currentBranch))
	if err != nil {
		return "", err
	}

	if len(revisionBoundaries) > 0 {
		return parseRefBoundary(revisionBoundaries), nil
	} else {
		currentRev, err := shell.Execute(fmt.Sprintf("git rev-parse %s", currentBranch))
		if err != nil {
			return "", err
		}

		cleanCurrentRev := strings.TrimSuffix(currentRev, "\n")
		return cleanCurrentRev, nil
	}
}

func Fetch() (string, error) {
	return shell.Execute("git fetch")
}

func ShowCommitMessage(branch string) (string, error) {
	return shell.Execute(fmt.Sprintf("git log origin/%s --format=%%B -n 1", branch))
}

func parseRefBoundary(revisionBoundaries string) string {
	boundaries := strings.FieldsFunc(revisionBoundaries, func(c rune) bool {
		return c == '\n'
	})

	return char.TrimFirstRune(boundaries[len(boundaries)-1])
}
