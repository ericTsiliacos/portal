package strategies

import "github.com/ericTsiliacos/portal/internal/portal/generator"

type Generate struct{}

func (gn Generate) Name() string {
	return "git-duet"
}

func (gn Generate) Strategy() string {
	return generator.RandomAdjective() + "-" + generator.RandomNoun()
}
