package main

import (
	"fmt"
	"os"
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

func commandFailure(command string, err error) {
	fmt.Println(command)
	fmt.Println("LogFile: " + logger.LogPath)
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
