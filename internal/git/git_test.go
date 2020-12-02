package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRefBoundary(t *testing.T) {
	revisionBoundaries := "4980d711afd8b8376d0404229bf1bb40b046247e\n-b90012997091b1dd3f2987f6495cc9b203fed291\n"
	actual := parseRefBoundary(revisionBoundaries)
	assert.Equal(t, actual, "b90012997091b1dd3f2987f6495cc9b203fed291")

	revisionBoundaries = "4980d711afd8b8376d0404229bf1bb40b046247e\n-b90012997091b1dd3f2987f6495cc9b203fed291"
	actual = parseRefBoundary(revisionBoundaries)
	assert.Equal(t, actual, "b90012997091b1dd3f2987f6495cc9b203fed291")
}
