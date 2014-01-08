package main

import (
	"fmt"
	"os"

	"github.com/flynn/go-discoverd"
	"github.com/flynn/lorne/client"
)

var cmdJobs = &Command{
	Run:   runJobs,
	Usage: "jobs",
	Short: "shows running cluster jobs",
	Long:  `Collects a list of jobs running on all hosts in the cluster.`,
}

func runJobs(cmd *Command, args []string) {
	if len(args) != 0 {
		cmd.printUsage()
		os.Exit(2)
	}
	err := discoverd.Connect(getTarget() + ":55002") // TODO: fix this
	assert(err)
	services, err := discoverd.Services("flynn-lorne", discoverd.DefaultTimeout)
	assert(err)
	for _, service := range services {
		host, err := client.New(service.Attrs["id"])
		assert(err)
		jobs, err := host.JobList()
		for k, _ := range jobs {
			fmt.Println(k, "\t", "["+service.Attrs["id"]+"]")
		}
	}
}
