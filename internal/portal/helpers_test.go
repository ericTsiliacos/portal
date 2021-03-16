package portal

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func SetupBareGitRepository(t *testing.T, rootDirectory string) {
	t.Helper()

	projectPath := filepath.Join(rootDirectory, "project")
	check(os.Mkdir(projectPath, 0755))

	check(os.Chdir(projectPath))

	_, err := exec.Command("git", "init", "--bare").Output()
	check(err)

	check(os.Chdir(rootDirectory))

	_, err = exec.Command("git", "clone", "project", "setup", "--progress").Output()
	check(err)

	setupPath := filepath.Join(rootDirectory, "setup")
	check(os.Chdir(setupPath))

	_, err = exec.Command("git", "commit", "--allow-empty", "--message", "\"initial commit\"").Output()
	check(err)

	_, err = exec.Command("git", "push", "origin", "head", "--progress").Output()
	check(err)

	check(os.Chdir(rootDirectory))
}

func CloneRepository(t *testing.T, rootDirectory string, name string) string {
	t.Helper()

	_, err := exec.Command("git", "clone", "project", name, "--progress").Output()
	check(err)

	clonePath := filepath.Join(rootDirectory, name)
	check(os.Chdir(clonePath))

	return clonePath
}

func RemoteBranchExists(t *testing.T, branch string) bool {
	t.Helper()

	remoteBranch, err := exec.Command("git", "ls-remote", "--heads", "origin", branch).Output()
	check(err)
	return len(remoteBranch) > 0
}

func LocalBranchExists(t *testing.T, branch string) bool {
	t.Helper()

	localBranch, err := exec.Command("git", "branch", "--list", branch).Output()
	check(err)

	return len(localBranch) > 0
}

func CleanIndex(t *testing.T) bool {
	t.Helper()

	index, err := exec.Command("git", "status", "--porcelain=v1").Output()
	check(err)

	indexCount := strings.Count(string(index), "\n")
	return !(indexCount > 0)
}
