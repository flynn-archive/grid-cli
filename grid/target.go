package main

import (
	"fmt"
	"os"
)

var cmdTarget = &Command{
	Run:   runTarget,
	Usage: "target <host>",
	Short: "target grid cluster",
	Long:  `Targets or shows target of a grid cluster using host in cluster.`,
}

func runTarget(cmd *Command, args []string) {
	if len(args) > 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	if len(args) == 1 {
		setTarget(args[0])
	} else {
		fmt.Println(getTarget())
	}
}
