package main

import (
	"fmt"
	"github.com/thatisuday/commando"
)

func main() {
	commando.
		SetExecutableName("portal").
		SetVersion("1.0.0").
		SetDescription("A commandline tool for moving work in progress to another machine using git")

	commando.
		Register("push").
		SetShortDescription("push work in progress").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			fmt.Println("Coming your way...")
		})

	commando.
		Register("pull").
		SetShortDescription("pull work in progress").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			fmt.Println("Coming your way...")
		})

	commando.Parse(nil)
}
