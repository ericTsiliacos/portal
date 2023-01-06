package portal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/saga"
	"github.com/stretchr/testify/assert"
)

func TestPortalPullSaga(t *testing.T) {
	portalBranch := "pa-ir-portal"
	fileName := "foo"

	currentBranch, sha := push(t, portalBranch, fileName)
	pullSteps, err := PullSagaSteps(context.TODO(), currentBranch, portalBranch, sha, false)
	if err != nil {
		t.FailNow()
	}
	pullSaga := saga.New(pullSteps)
	errors := pullSaga.Run()

	assert.Empty(t, errors)
	assert.FileExists(t, fileName)
	assert.False(t, RemoteBranchExists(t, portalBranch))
	assert.False(t, LocalBranchExists(t, portalBranch))
	assert.False(t, CleanIndex(t))
}

func TestPortalPullSagaWithFailures(t *testing.T) {
	portalBranch := "pa-ir-portal"
	fileName := "foo"

	currentBranch, sha := push(t, portalBranch, fileName)

	pullSteps, err := PullSagaSteps(context.TODO(), currentBranch, portalBranch, sha, false)
	if err != nil {
		t.FailNow()
	}

	for i := 1; i < len(pullSteps); i++ {
		fmt.Printf("test run %d:", i)
		fmt.Println()
		fmt.Println("-----------------")

		testPortalPullSagaFailure(t, portalBranch, i)
	}
}

func testPortalPullSagaFailure(t *testing.T, portalBranch string, index int) {
	rootDirectory := t.TempDir()

	SetupBareGitRepository(t, rootDirectory)

	clone1Path := CloneRepository(t, rootDirectory, "clone1")

	check(os.Chdir(rootDirectory))
	check(os.Chdir(CloneRepository(t, rootDirectory, "clone2")))
	fileName := "foo"
	fileHandle, err := os.Create(fileName)
	check(err)
	defer fileHandle.Close()

	pushSteps, err := PushSagaSteps(context.TODO(), portalBranch, "v1.0.0", false, "")
	if err != nil {
		t.FailNow()
	}
	pushSaga := saga.New(pushSteps)
	errs := pushSaga.Run()
	assert.Empty(t, errs)

	remoteTrackingBranch, _ := git.GetRemoteTrackingBranch()
	currentBranch, _ := git.GetCurrentBranch()
	sha, _ := git.GetBoundarySha(remoteTrackingBranch, currentBranch)

	check(os.Chdir(clone1Path))
	git.Fetch()
	pullSteps, _ := PullSagaSteps(context.TODO(), currentBranch, portalBranch, sha, false)
	if err != nil {
		t.FailNow()
	}
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

func push(t *testing.T, portalBranch string, fileName string) (string, string) {
	rootDirectory := t.TempDir()

	SetupBareGitRepository(t, rootDirectory)

	clone1Path := CloneRepository(t, rootDirectory, "clone1")

	check(os.Chdir(rootDirectory))
	check(os.Chdir(CloneRepository(t, rootDirectory, "clone2")))
	fileHandle, err := os.Create(fileName)
	check(err)
	defer fileHandle.Close()

	pushSteps, err := PushSagaSteps(context.TODO(), portalBranch, "v1.0.0", false, "")
	if err != nil {
		t.FailNow()
	}
	pushSaga := saga.New(pushSteps)
	errors := pushSaga.Run()
	assert.Empty(t, errors)

	remoteTrackingBranch, err := git.GetRemoteTrackingBranch()
	if err != nil {
		t.FailNow()
	}
	currentBranch, err := git.GetCurrentBranch()
	if err != nil {
		t.FailNow()
	}
	sha, err := git.GetBoundarySha(remoteTrackingBranch, currentBranch)
	if err != nil {
		t.FailNow()
	}

	check(os.Chdir(clone1Path))
	git.Fetch()

	return currentBranch, sha
}
