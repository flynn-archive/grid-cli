package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/flynn/flynn-host/types"
	"github.com/flynn/go-discoverd"
	"github.com/flynn/go-dockerclient"
	"github.com/flynn/go-flynn/cluster"
)

var cmdSchedule = &Command{
	Run:   runSchedule,
	Usage: "schedule <image>",
	Short: "schedules a job to run",
	Long:  `Schedules a (service) job to run by image name.`,
}

func runSchedule(cmd *Command, args []string) {
	if len(args) != 1 {
		cmd.printUsage()
		os.Exit(2)
	}
	err := discoverd.Connect(getTarget() + ":55002") // TODO: fix this
	assert(err)
	client, err := cluster.NewClient()
	assert(err)

	hosts, err := client.ListHosts()
	assert(err)
	var hostid string
	for id, _ := range hosts {
		hostid = id
		break
	}

	jobid := cluster.RandomJobID(args[0])
	config := docker.Config{
		Image: args[0],
		//Cmd:          []string{"start", "web"},
		Tty:          false,
		AttachStdin:  false,
		AttachStdout: false,
		AttachStderr: false,
		OpenStdin:    false,
		StdinOnce:    false,
	}

	jobReq := &host.AddJobsReq{
		Incremental: true,
		HostJobs: map[string][]*host.Job{
			hostid: {{ID: jobid, Config: &config, TCPPorts: 1}}},
	}
	_, err = client.AddJobs(jobReq)
	assert(err)

	if addr := getAddr(client, hostid, jobid); addr != "" {
		fmt.Println(jobid + " created, listening at " + addr)
	} else {
		fmt.Println("Service was not scheduled")
		os.Exit(1)
	}
}

func getAddr(client *cluster.Client, hostid, jobid string) string {
	services, err := discoverd.Services("flynn-host", discoverd.DefaultTimeout)
	var service *discoverd.Service
	for _, s := range services {
		if s.Attrs["id"] == hostid {
			service = s
			break
		}
	}
	if service == nil {
		return ""
	}
	host, err := client.ConnectHost(hostid)
	assert(err)
	job, err := host.GetJob(jobid)
	assert(err)
	for portspec := range job.Job.Config.ExposedPorts {
		return service.Host + ":" + strings.Split(portspec, "/")[0]
	}
	return ""
}
