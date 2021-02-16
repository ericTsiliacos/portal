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

func CurrentWorkingDirectoryGitRoot() bool {
	gitRoot := strings.TrimSuffix(shell.Check(shell.Execute("git rev-parse --git-dir")), "\n")

	return gitRoot == ".git"
}

func Version() string {
	return shell.Check(shell.Execute("git version"))
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

func BuildPatch(remoteTrackingBranch string) (string, error) {
	return shell.Execute(fmt.Sprintf("git format-patch %s --stdout", remoteTrackingBranch))
}

func GetCurrentBranch() string {
	currentBranch, _ := shell.Execute("git rev-parse --abbrev-ref HEAD")
	cleanCurrentBranch := strings.TrimSuffix(currentBranch, "\n")
	return cleanCurrentBranch
}

func GetRemoteTrackingBranch() string {
	remoteTrackingBranch, _ := shell.Execute("git rev-parse --abbrev-ref --symbolic-full-name @{u}")
	cleanRemoteTrackingBranch := strings.TrimSuffix(remoteTrackingBranch, "\n")
	return cleanRemoteTrackingBranch
}

func GetBoundarySha(remoteTrackingBranch string, currentBranch string) string {
	revisionBoundaries, _ := shell.Execute(fmt.Sprintf("git rev-list --boundary %s..%s", remoteTrackingBranch, currentBranch))
	if len(revisionBoundaries) > 0 {
		return parseRefBoundary(revisionBoundaries)
	} else {
		currentRev, _ := shell.Execute(fmt.Sprintf("git rev-parse %s", currentBranch))
		cleanCurrentRev := strings.TrimSuffix(currentRev, "\n")
		return cleanCurrentRev
	}
}

func Add(files string) (string, error) {
	_, err := shell.Execute(fmt.Sprintf("git add %s", files))
	return files, err
}

func Reset() (string, error) {
	return shell.Execute("git reset")
}

func UndoCommit() (string, error) {
	return shell.Execute("git reset HEAD^")
}

func Commit(message string) (string, error) {
	return shell.Execute(fmt.Sprintf("git commit --allow-empty -m %s", message))
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
