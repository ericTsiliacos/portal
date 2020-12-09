package shell

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ericTsiliacos/portal/internal/logger"
)

func Execute(command string) (string, error) {
	logger.LogInfo.Println(command)

	s := strings.Split(command, " ")
	cmd, args := s[0], s[1:]
	cmdOut, err := exec.Command(cmd, args...).CombinedOutput()
	output := string(cmdOut)

	logger.LogInfo.Println(output)

	if err != nil {
		logger.LogError.Println(err)
	}

	return output, err
}

func Check(output string, err error) string {
	if err != nil {
		fmt.Println()
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return output
}
