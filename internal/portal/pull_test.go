package portal

import (
	"log"
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

	log.Println(remoteTrackingBranch)
	log.Println(currentBranch)
	log.Println(sha)

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
