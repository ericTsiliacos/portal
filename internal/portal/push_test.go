package portal

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ericTsiliacos/portal/internal/saga"
	"github.com/stretchr/testify/assert"
)

func TestPortalPushSaga(t *testing.T) {
	rootDirectory := t.TempDir()

	SetupBareGitProject(t, rootDirectory)

	check(os.Chdir(SetupProjectClone(t, rootDirectory)))

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
	assert.True(t, RemoteBranchExists(t, portalBranch))
	assert.False(t, LocalBranchExists(t, portalBranch))
	assert.True(t, CleanIndex(t))
}

func TestPortalPushSagaWithFailures(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	portalBranch := "pa-ir-portal"
	steps := PushSagaSteps(portalBranch, now, "v1.0.0", true)

	for i := 1; i < len(steps); i++ {
		fmt.Printf("test run %d:", i)
		fmt.Println()
		fmt.Println("-----------------")

		testPortalPushSagaFailure(t, portalBranch, now, i)
	}
}

func testPortalPushSagaFailure(t *testing.T, portalBranch string, now string, index int) {
	rootDirectory := t.TempDir()

	SetupBareGitProject(t, rootDirectory)

	check(os.Chdir(SetupProjectClone(t, rootDirectory)))

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
	assert.False(t, RemoteBranchExists(t, portalBranch))
	assert.False(t, LocalBranchExists(t, portalBranch))
	assert.False(t, CleanIndex(t))
}
