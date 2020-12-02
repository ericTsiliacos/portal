package fs

import (
	"fmt"

	"github.com/ericTsiliacos/portal/internal/shell"
)

func removeFile(filename string) (string, error) {
	return shell.Execute(fmt.Sprintf("rm %s", filename))
}
