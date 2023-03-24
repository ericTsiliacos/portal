package portal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ericTsiliacos/portal/internal/saga"
)

func TestPortalPushSaga(t *testing.T) {
	fileName := "foo"
	portalBranch := "pa-ir-portal"

	pushSetup(t, fileName)

	steps, err := PushSagaSteps(context.TODO(), portalBranch, "v1.0.0", false, "")
	if err != nil {
		t.FailNow()
	}
	saga := saga.New(steps)
	errors := saga.Run()

	assert.Empty(t, errors)
	assert.NoFileExists(t, fileName)
	assert.True(t, RemoteBranchExists(t, portalBranch))
	assert.False(t, LocalBranchExists(t, portalBranch))
	assert.True(t, CleanIndex(t))
}

func TestPortalPushSagaWithFailures(t *testing.T) {
	fileName := "foo"
	portalBranch := "pa-ir-portal"

	pushSetup(t, fileName)
	steps, err := PushSagaSteps(context.TODO(), portalBranch, "v1.0.0", false, "")
	if err != nil {
		t.FailNow()
	}

	for i := 1; i < len(steps); i++ {
		fmt.Printf("test run %d:", i)
		fmt.Println()
		fmt.Println("-----------------")

		testPortalPushSagaFailure(t, portalBranch, i)
	}
}

func testPortalPushSagaFailure(t *testing.T, portalBranch string, index int) {
	rootDirectory := t.TempDir()

	SetupBareGitRepository(t, rootDirectory)

	check(os.Chdir(CloneRepository(t, rootDirectory, "clone1")))

	fileName := "foo"
	fileHandle, err := os.Create(fileName)
	check(err)
	defer fileHandle.Close()

	steps, err := PushSagaSteps(context.TODO(), portalBranch, "v1.0.0", false, "")
	if err != nil {
		t.FailNow()
	}
	steps = steps[0 : len(steps)-index]
	stepsWithError := append(steps, saga.Step{
		Name: "Boom!",
		Run: func() error {
			return errors.New("uh oh!")
		},
	})
	saga := saga.New(stepsWithError)
	errors := saga.Run()

	assert.NotEmpty(t, errors)
	assert.FileExists(t, fileName)
	assert.False(t, RemoteBranchExists(t, portalBranch))
	assert.False(t, LocalBranchExists(t, portalBranch))
	assert.False(t, CleanIndex(t))
}

func pushSetup(t *testing.T, fileName string) {
	rootDirectory := t.TempDir()

	SetupBareGitRepository(t, rootDirectory)

	check(os.Chdir(CloneRepository(t, rootDirectory, "clone1")))

	fileHandle, err := os.Create(fileName)
	check(err)
	defer fileHandle.Close()
}
