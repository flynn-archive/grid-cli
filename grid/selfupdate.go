package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"

	"bitbucket.org/kardianos/osext"
)

// TODO: checksum downloads

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
	resp, err := http.Get("https://s3.amazonaws.com/progrium-flynn/flynn-grid/dev/version")
	assert(err)
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	assert(err)
	version := string(body)
	if version == "dev" {
		updateSelf()
		return
	}
	if version == Version {
		fmt.Println("Already up to date.")
	}
	updateSelf()
}

func updateSelf() {
	resp, err := http.Get("https://s3.amazonaws.com/progrium-flynn/flynn-grid/dev/grid-cli_" + runtime.GOOS + "_" + runtime.GOARCH + ".tgz")
	assert(err)
	defer resp.Body.Close()
	selfpath, err := osext.Executable()
	assert(err)
	info, err := os.Stat(selfpath)
	assert(err)
	f, err := ioutil.TempFile("", "grid-update")
	assert(err)
	_, err = io.Copy(f, resp.Body)
	assert(err)
	f.Close()
	assert(f.Chmod(info.Mode().Perm()))
	assert(os.Remove(selfpath))
	assert(os.Rename(f.Name(), selfpath))
	fmt.Println("Updated.")
}
