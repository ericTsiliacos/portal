package main

import (
	"sort"
	"strings"
)

func branchName() string {
	author := execute("git config --get duet.env.git-author-initials")
	coauthor := execute("git config --get duet.env.git-committer-initials")

	authors := []string{author, coauthor}
	sort.Strings(authors)
	authors = Map(authors, func(s string) string {
		return strings.TrimSuffix(s, "\n")
	})
	branch := strings.Join(authors, "-")
	return branch
}
