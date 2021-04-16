package portal

import (
	"fmt"
	"os/exec"

	"github.com/ericTsiliacos/portal/internal/logger"
)

func Run(cmd *exec.Cmd, verbose bool) (err error) {
	if verbose {
		fmt.Println(cmd.String())
	}

	logger.LogInfo.Println(cmd.String())
	err = cmd.Run()
	return
}
