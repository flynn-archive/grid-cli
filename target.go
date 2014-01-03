package main

import (
	"fmt"
	"os"
)

var cmdTarget = &Command{
	Run:   runTarget,
	Usage: "target <ip>",
	Short: "target grid cluster",
	Long:  `Targets a grid cluster using IP of host in cluster.`,
}

func runTarget(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.printUsage()
		os.Exit(2)
	}
	fmt.Println("Target!")
}
