package main

import (
	"os/exec"
	"strings"

	"github.com/ericTsiliacos/portal/logger"
)

func execute(command string) (string, error) {
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
