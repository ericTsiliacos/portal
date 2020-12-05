package strategies

import (
	"sort"
	"strings"

	"github.com/ericTsiliacos/portal/internal/slices"
)

type Strategy interface {
	Name() string
	Strategy() string
}

func getAuthorsBranch(authors []string) string {
	if len(authors) > 0 {
		sort.Strings(authors)
		authors = slices.Map(authors, func(s string) string {
			return strings.TrimSuffix(s, "\n")
		})
		branch := strings.Join(authors, "-")
		return prefixPortal(branch)
	} else {
		return ""
	}
}

func prefixPortal(branchName string) string {
	return "portal-" + branchName
}
