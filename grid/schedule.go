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

	hostClient, err := client.ConnectHost(hostid)
	assert(err)

	events := make(chan *host.Event)
	addr := make(chan string)
	hostClient.StreamEvents(jobid, events)
	go func() {
		for event := range events {
			switch event.Event {
			case "start":
				addr <- getAddr(hostClient, hostid, jobid)
				return
			case "error", "stop":
				fmt.Println("Scheduling error")
				// TODO: read error from host
				os.Exit(1)
			}
		}
	}()
	_, err = client.AddJobs(jobReq)
	assert(err)

	if a := <-addr; a != "" {
		fmt.Println(jobid + " created, listening at " + a)
	} else {
		fmt.Println("Service was scheduled but no exposed port found")
		os.Exit(1)
	}

}

func getAddr(host cluster.Host, hostid, jobid string) string {
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
	job, err := host.GetJob(jobid)
	assert(err)
	for portspec := range job.Job.Config.ExposedPorts {
		return service.Host + ":" + strings.Split(portspec, "/")[0]
	}
	return ""
}
