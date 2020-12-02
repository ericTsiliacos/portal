package portal

import (
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/ericTsiliacos/portal/internal/git"
	"github.com/ericTsiliacos/portal/internal/slices"
	"gopkg.in/yaml.v2"
)

type Meta struct {
	Meta struct {
		Version       string `yaml:"version"`
		WorkingBranch string `yaml:"workingBranch"`
		Sha           string `yaml:"sha"`
	} `yaml:"Meta"`
}

func GetPortalBranch(authors []string) string {
	sort.Strings(authors)
	authors = slices.Map(authors, func(s string) string {
		return strings.TrimSuffix(s, "\n")
	})
	branch := strings.Join(authors, "-")
	return portal(branch)
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

func BranchNameStrategy(strategy string) ([]string, error) {
	strategies := map[string]interface{}{
		"git-duet":     git.GitDuet,
		"git-together": git.GitTogether,
	}

	if strategy != "auto" {
		fn, ok := strategies[strategy]
		if !ok {
			return nil, errors.New("unknown strategy")
		} else {
			return fn.(func() []string)(), nil
		}
	}

	pairs := [][]string{
		git.GitDuet(),
		git.GitTogether(),
	}

	if slices.All(pairs, slices.Empty) {
		return nil, errors.New("no branch naming strategy found")
	}

	if slices.Many(pairs, slices.NonEmpty) {
		return nil, errors.New("multiple branch naming strategies found")
	}

	return slices.FindFirst(pairs, slices.NonEmpty), nil
}

func portal(branchName string) string {
	return "portal-" + branchName
}

func buildPatchFileName(dateTime string) string {
	return fmt.Sprintf("portal-%s.patch", dateTime)
}
