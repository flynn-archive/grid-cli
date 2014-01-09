package main

import (
	"fmt"
	"os"
)

var cmdSelfUpdate = &Command{
	Run:   runSelfUpdate,
	Usage: "selfupdate",
	Short: "update the grid tool",
	Long:  `Checks for an update and downloads a new version if available`,
}

func runSelfUpdate(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.printUsage()
		os.Exit(2)
	}
	fmt.Println("selfupdate!")
}
