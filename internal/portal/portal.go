package portal

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/portal/strategies"
	"github.com/ericTsiliacos/portal/internal/saga"
	"github.com/ericTsiliacos/portal/internal/shell"
	"gopkg.in/yaml.v2"
)

func PushSagaSteps(portalBranch string, now string, version string, patch bool) []saga.Step {
	remoteTrackingBranch := git.GetRemoteTrackingBranch()
	currentBranch := git.GetCurrentBranch()
	sha := git.GetBoundarySha(remoteTrackingBranch, currentBranch)

	return []saga.Step{
		{
			Name:       "git add .",
			Run:        func() (string, error) { return git.Add(".") },
			Compensate: func(input string) (string, error) { return git.Reset() },
		},
		{
			Name:       "git commit -m 'portal-wip'",
			Run:        func() (string, error) { return git.Commit("portal-wip") },
			Compensate: func(input string) (string, error) { return git.UndoCommit() },
		},
		{
			Name: "create git patch of work in progress",
			Run: func() (string, error) {
				return Patch(remoteTrackingBranch, now)
			},
			Compensate: func(fileName string) (string, error) {
				return shell.Execute(fmt.Sprintf("rm %s", fileName))
			},
			Exclude: !patch,
		},
		{
			Name: "create portal-meta.yml",
			Run: func() (string, error) {
				return WritePortalMetadata("portal-meta.yml", currentBranch, sha, version)
			},
			Compensate: func(fileName string) (string, error) {
				return shell.Execute(fmt.Sprintf("rm %s", fileName))
			},
		},
		{
			Name:       "git add portal-meta.yml",
			Run:        func() (string, error) { return git.Add("portal-meta.yml") },
			Compensate: func(input string) (string, error) { return git.Reset() },
		},
		{
			Name:       "git commit -m 'portal-meta'",
			Run:        func() (string, error) { return git.Commit("portal-meta") },
			Compensate: func(input string) (string, error) { return git.UndoCommit() },
		},
		{
			Name: "git stash backup patch",
			Run: func() (string, error) {
				return shell.Execute(fmt.Sprintf("git stash push -m \"portal-patch-%s\" --include-untracked", now))
			},
			Compensate: func(input string) (string, error) {
				return shell.Execute("git stash pop")
			},
			Exclude: !patch,
		},
		{
			Name: "git checkout portal branch",
			Run: func() (string, error) {
				_, err := shell.Execute(fmt.Sprintf("git checkout -b %s --progress", portalBranch))
				return currentBranch, err
			},
			Compensate: func(currentBranch string) (string, error) {
				shell.Execute(fmt.Sprintf("git checkout %s --progress", currentBranch))

				return shell.Execute(fmt.Sprintf("git branch -D %s", portalBranch))
			},
		},
		{
			Name: "git push portal branch",
			Run: func() (string, error) {
				return shell.Execute(fmt.Sprintf("git push origin %s --progress", portalBranch))
			},
			Compensate: func(originalBranch string) (string, error) {
				return shell.Execute(fmt.Sprintf("git push origin --delete %s --progress", portalBranch))
			},
			Retries: 1,
		},
		{
			Name: "git checkout to original branch",
			Run: func() (string, error) {
				return shell.Execute("git checkout - --progress")
			},
		},
		{
			Name: "delete local portal branch",
			Run: func() (string, error) {
				return shell.Execute(fmt.Sprintf("git branch -D %s", portalBranch))
			},
			Compensate: func(originalBranch string) (string, error) {
				return shell.Execute(fmt.Sprintf("git checkout -b %s --progress", portalBranch))
			},
		},
		{
			Name: "clear git workspace",
			Run: func() (string, error) {
				return shell.Execute(fmt.Sprintf("git reset --hard %s", remoteTrackingBranch))
			},
		},
	}
}

type Meta struct {
	Meta struct {
		Version       string `yaml:"version"`
		WorkingBranch string `yaml:"workingBranch"`
		Sha           string `yaml:"sha"`
	} `yaml:"Meta"`
}

func GetConfiguration(yamlContent string) (*Meta, error) {
	c := &Meta{}
	err := yaml.Unmarshal([]byte(yamlContent), c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c, err
}

func WritePortalMetadata(fileName string, branch string, sha string, version string) (string, error) {
	f, err := os.Create(fileName)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	c := Meta{}
	c.Meta.WorkingBranch = branch
	c.Meta.Sha = sha
	c.Meta.Version = version

	d, marshalError := yaml.Marshal(&c)
	if marshalError != nil {
		log.Fatalf("error: %v", marshalError)
	}

	_, err = f.WriteString(string(d))
	if err != nil {
		fmt.Println(err)
		_ = f.Close()

		return "", err
	}

	return fileName, nil
}

func Patch(remoteTrackingBranch string, dateTime string) (string, error) {
	patch, err := git.BuildPatch(remoteTrackingBranch)
	if err != nil {
		return "", err
	}
	fileName := buildPatchFileName(dateTime)
	f, err := os.Create(fileName)
	if err != nil {
		return "", err
	}

	_, err = f.WriteString(patch)
	if err != nil {
		return "", err
	}

	err = f.Close()
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func BranchNameStrategy(strategyName string) (string, error) {
	strategies := []strategies.Strategy{
		strategies.GitDuet{},
		strategies.GitTogether{},
	}

	if strategyName != "auto" {
		for _, strategy := range strategies {
			if strategy.Name() == strategyName {
				branchName := strategy.Strategy()
				if branchName != "" {
					return branchName, nil
				} else {
					return "", fmt.Errorf("%s not configured", strategy.Name())
				}
			}
		}

		return "", errors.New("unknown strategy")
	}

	branchNames := []string{}
	for i := 0; i < len(strategies); i++ {
		if strategies[i].Strategy() != "" {
			branchNames = append(branchNames, strategies[i].Strategy())
		}
	}

	if len(branchNames) == 0 {
		return "", errors.New("no branch naming strategy found")
	}

	if len(branchNames) > 1 {
		return "", errors.New("multiple branch naming strategies found")
	}

	return branchNames[0], nil
}

func buildPatchFileName(dateTime string) string {
	return fmt.Sprintf("portal-%s.patch", dateTime)
}
