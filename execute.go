package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func execute(command string) string {
	s := strings.Split(command, " ")
	cmd, args := s[0], s[1:]
	cmdOut, err := exec.Command(cmd, args...).Output()

	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	return string(cmdOut)
}
