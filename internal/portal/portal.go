package portal

import (
	"errors"
	"fmt"
	"log"

	"github.com/ericTsiliacos/portal/internal/portal/strategies"
	"gopkg.in/yaml.v2"
)

type Meta struct {
	Meta struct {
		Version       string `yaml:"version"`
		WorkingBranch string `yaml:"workingBranch"`
		Sha           string `yaml:"sha"`
		MsgPostfix	  string `yaml:"msgPostfix"`
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
