package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func execute(command string) (string, error) {
	s := strings.Split(command, " ")
	cmd, args := s[0], s[1:]
	cmdOut, err := exec.Command(cmd, args...).CombinedOutput()
	return string(cmdOut), err
}

func commandFailure(command string, err error) {
	fmt.Println(command)
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
