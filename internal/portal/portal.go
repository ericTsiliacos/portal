package portal

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/portal/strategies"
	"gopkg.in/yaml.v2"
)

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

func WritePortalMetaData(branch string, sha string, version string) {
	f, err := os.Create("portal-meta.yml")
	if err != nil {
		fmt.Println(err)
		return
	}

	c := Meta{}
	c.Meta.WorkingBranch = branch
	c.Meta.Sha = sha
	c.Meta.Version = version

	d, err := yaml.Marshal(&c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	_, err = f.WriteString(string(d))
	if err != nil {
		fmt.Println(err)
		_ = f.Close()
		return
	}

}

func SavePatch(remoteTrackingBranch string, dateTime string) {
	patch, _ := git.BuildPatch(remoteTrackingBranch)
	f, _ := os.Create(buildPatchFileName(dateTime))
	_, _ = f.WriteString(patch)
	_ = f.Close()
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
