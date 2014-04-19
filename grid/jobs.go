package main

import (
	"fmt"
	"github.com/flynn/go-flynn/cluster"
	"os"
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
	client, err := cluster.NewClient()
	assert(err)
	hosts, err := client.ListHosts()
	assert(err)
	for hostid, _ := range hosts {
		host, err := client.DialHost(hostid)
		assert(err)
		jobs, err := host.ListJobs()
		assert(err)
		for jobid, _ := range jobs {
			fmt.Println(jobid, "\t", "["+hostid+"]")
		}
	}
}
