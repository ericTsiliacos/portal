package main

import "strings"

func gitTogether() []string {
	activeAuthors := execute("git config --get git-together.active")
	return strings.Split(activeAuthors, "+")
}
