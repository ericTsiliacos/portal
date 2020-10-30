package main

import (
	"os/exec"
	"strings"
)

func execute(command string) (string, error) {
	s := strings.Split(command, " ")
	cmd, args := s[0], s[1:]
	cmdOut, err := exec.Command(cmd, args...).Output()

	return string(cmdOut), err
}
