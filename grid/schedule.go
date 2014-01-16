package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/flynn/go-discoverd"
	"github.com/flynn/go-dockerclient"
	hosts "github.com/flynn/lorne/client"
	"github.com/flynn/sampi/client"
	"github.com/flynn/sampi/types"
	"github.com/nu7hatch/gouuid"
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
	sched, err := client.New()
	assert(err)

	hosts, err := discoverd.Services("flynn-lorne", discoverd.DefaultTimeout)
	assert(err)
	hostid := hosts[0].Attrs["id"]

	id, err := uuid.NewV4()
	jobid := id.String()
	assert(err)
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

	schedReq := &sampi.ScheduleReq{
		Incremental: true,
		HostJobs: map[string][]*sampi.Job{
			hostid: {{ID: jobid, Config: &config, TCPPorts: 1}}},
	}
	_, err = sched.Schedule(schedReq)
	assert(err)

	if port := getPort(hostid, jobid); port != "" {
		fmt.Println(jobid + " created, listening at " + hosts[0].Host + ":" + port)
	} else {
		fmt.Println("Service was not scheduled")
		os.Exit(1)
	}
}

func getPort(hostid, jobid string) string {
	host, err := hosts.New(hostid)
	assert(err)
	job, err := host.GetJob(jobid)
	assert(err)
	for portspec := range job.Job.Config.ExposedPorts {
		return strings.Split(portspec, "/")[0]
	}
	return ""
}
