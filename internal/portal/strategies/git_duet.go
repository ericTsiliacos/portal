package strategies

import (
	"github.com/ericTsiliacos/portal/internal/git"
)

type GitDuet struct{}

func (gd GitDuet) Name() string {
	return "git-duet"
}

func (gd GitDuet) Strategy() string {
	return getAuthorsBranch(git.GitDuet())
}
