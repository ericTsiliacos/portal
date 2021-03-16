package portal

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/saga"
	"github.com/stretchr/testify/assert"
)

func TestPortalPullSaga(t *testing.T) {
	rootDirectory := t.TempDir()

	SetupBareGitRepository(t, rootDirectory)

	clone1Path := CloneRepository(t, rootDirectory, "clone1")

	check(os.Chdir(rootDirectory))
	check(os.Chdir(CloneRepository(t, rootDirectory, "clone2")))
	fileName := "foo"
	fileHandle, err := os.Create(fileName)
	check(err)
	defer fileHandle.Close()

	now := time.Now().Format(time.RFC3339)
	portalBranch := "pa-ir-portal"
	pushSteps := PushSagaSteps(portalBranch, now, "v1.0.0")
	pushSaga := saga.New(pushSteps)
	errors := pushSaga.Run()
	assert.Empty(t, errors)

	remoteTrackingBranch := git.GetRemoteTrackingBranch()
	currentBranch := git.GetCurrentBranch()
	sha := git.GetBoundarySha(remoteTrackingBranch, currentBranch)

	check(os.Chdir(clone1Path))
	git.Fetch()
	pullSteps := PullSagaSteps(currentBranch, portalBranch, sha)
	pullSaga := saga.New(pullSteps)
	errors = pullSaga.Run()

	assert.Empty(t, errors)
	assert.FileExists(t, fileName)
	assert.False(t, RemoteBranchExists(t, portalBranch))
	assert.False(t, LocalBranchExists(t, portalBranch))
	assert.False(t, CleanIndex(t))
}

func TestPortalPullSagaWithFailures(t *testing.T) {
	now := time.Now().Format(time.RFC3339)
	portalBranch := "pa-ir-portal"
	pullSteps := PullSagaSteps("", "", "")

	for i := 1; i < len(pullSteps); i++ {
		fmt.Printf("test run %d:", i)
		fmt.Println()
		fmt.Println("-----------------")

		testPortalPullSagaFailure(t, portalBranch, now, i)
	}
}

func testPortalPullSagaFailure(t *testing.T, portalBranch string, now string, index int) {
	rootDirectory := t.TempDir()

	SetupBareGitRepository(t, rootDirectory)

	clone1Path := CloneRepository(t, rootDirectory, "clone1")

	check(os.Chdir(rootDirectory))
	check(os.Chdir(CloneRepository(t, rootDirectory, "clone2")))
	fileName := "foo"
	fileHandle, err := os.Create(fileName)
	check(err)
	defer fileHandle.Close()

	pushSteps := PushSagaSteps(portalBranch, now, "v1.0.0")
	pushSaga := saga.New(pushSteps)
	errs := pushSaga.Run()
	assert.Empty(t, errs)

	remoteTrackingBranch := git.GetRemoteTrackingBranch()
	currentBranch := git.GetCurrentBranch()
	sha := git.GetBoundarySha(remoteTrackingBranch, currentBranch)

	check(os.Chdir(clone1Path))
	git.Fetch()
	pullSteps := PullSagaSteps(currentBranch, portalBranch, sha)
	pullSteps = pullSteps[0 : len(pullSteps)-index]
	stepsWithError := append(pullSteps, saga.Step{
		Name: "Boom!",
		Run: func() error {
			return errors.New("uh oh!")
		},
	})
	pullSaga := saga.New(stepsWithError)
	errs = pullSaga.Run()

	assert.NotEmpty(t, errs)
	assert.NoFileExists(t, fileName)
	assert.True(t, RemoteBranchExists(t, portalBranch))
	assert.False(t, LocalBranchExists(t, portalBranch))
	assert.True(t, CleanIndex(t))
}
