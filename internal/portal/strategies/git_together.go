package strategies

import (
	"github.com/ericTsiliacos/portal/internal/git"
)

type GitTogether struct{}

func (gt GitTogether) Name() string {
	return "git-together"
}

func (gt GitTogether) Strategy() string {
	return getAuthorsBranch(git.GitTogether())
}
