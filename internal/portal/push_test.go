package portal

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ericTsiliacos/portal/internal/saga"
	"github.com/stretchr/testify/assert"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func setupBareGitProject(t *testing.T, rootDirectory string) {
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

func setupProjectClone(t *testing.T, rootDirectory string) string {
	t.Helper()

	_, err := exec.Command("git", "clone", "project", "clone1", "--progress").Output()
	check(err)

	clonePath := filepath.Join(rootDirectory, "clone1")
	check(os.Chdir(clonePath))

	return clonePath
}

func remoteBranchExists(t *testing.T, branch string) bool {
	t.Helper()

	remoteBranch, err := exec.Command("git", "ls-remote", "--heads", "origin", branch).Output()
	check(err)
	return len(remoteBranch) > 0
}

func localBranchExists(t *testing.T, branch string) bool {
	t.Helper()

	localBranch, err := exec.Command("git", "branch", "--list", branch).Output()
	check(err)

	return len(localBranch) > 0
}

func cleanIndex(t *testing.T) bool {
	t.Helper()

	index, err := exec.Command("git", "status", "--porcelain=v1").Output()
	check(err)

	indexCount := strings.Count(string(index), "\n")
	return !(indexCount > 0)
}

func testSagaFailure(t *testing.T, portalBranch string, now string, index int) {
	rootDirectory := t.TempDir()

	setupBareGitProject(t, rootDirectory)

	check(os.Chdir(setupProjectClone(t, rootDirectory)))

	fileName := "foo"
	fileHandle, err := os.Create(fileName)
	check(err)
	defer fileHandle.Close()

	steps := PushSagaSteps(portalBranch, now, "v1.0.0", true)
	steps = steps[0 : len(steps)-index]
	stepsWithError := append(steps, saga.Step{
		Name: "Boom!",
		Run: func() (string, error) {
			return "", errors.New("uh oh!")
		},
	})

	saga := saga.Saga{Steps: stepsWithError, Verbose: testing.Verbose()}
	errors := saga.Run()

	assert.NotEmpty(t, errors)
	assert.FileExists(t, fileName)
	assert.False(t, remoteBranchExists(t, portalBranch))
	assert.False(t, localBranchExists(t, portalBranch))
	assert.False(t, cleanIndex(t))
}

func TestPortalPushSaga(t *testing.T) {
	rootDirectory := t.TempDir()

	setupBareGitProject(t, rootDirectory)

	check(os.Chdir(setupProjectClone(t, rootDirectory)))

	fileName := "foo"
	fileHandle, err := os.Create(fileName)
	check(err)
	defer fileHandle.Close()

	now := time.Now().Format(time.RFC3339)
	portalBranch := "pa-ir-portal"
	steps := PushSagaSteps(portalBranch, now, "v1.0.0", true)
	saga := saga.Saga{Steps: steps, Verbose: testing.Verbose()}
	errors := saga.Run()

	assert.Empty(t, errors)
	assert.NoFileExists(t, fileName)
	assert.True(t, remoteBranchExists(t, portalBranch))
	assert.False(t, localBranchExists(t, portalBranch))
	assert.True(t, cleanIndex(t))
}

func TestPortalPushSagaWithFailures(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	portalBranch := "pa-ir-portal"
	steps := PushSagaSteps(portalBranch, now, "v1.0.0", true)

	for i := 1; i < len(steps); i++ {
		fmt.Printf("test run %d:", i)
		fmt.Println()
		fmt.Println("-----------------")

		testSagaFailure(t, portalBranch, now, i)
	}
}
