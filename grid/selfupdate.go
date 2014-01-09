package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"bitbucket.org/kardianos/osext"
)

// TODO: checksum downloads

var showUpdateNoticeCh = make(chan bool, 1)

func init() {
	if Version == "dev" || (len(os.Args) > 1 && os.Args[1] == "selfupdate") {
		showUpdateNoticeCh <- false
		return
	}
	go func() {
		shouldUpdate, err := checkForUpdate()
		if err == nil && shouldUpdate {
			showUpdateNoticeCh <- true
			return
		}
		showUpdateNoticeCh <- false
	}()
}

func showUpdateNotice() {
	select {
	case show := <-showUpdateNoticeCh:
		if show {
			fmt.Fprintln(os.Stderr, "\033[32mNew version is available. Run `grid selfupdate` to get it. \033[39m")
		}
	case <-time.After(time.Second):
	}
}

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
	if Version != "dev" {
		shouldUpdate, err := checkForUpdate()
		assert(err)
		if !shouldUpdate {
			fmt.Println("Already up to date.")
			return
		}
	}
	updateSelf()
}

func checkForUpdate() (bool, error) {
	resp, err := http.Get("https://s3.amazonaws.com/progrium-flynn/flynn-grid/dev/version")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(string(body)) != Version {
		return true, nil
	}
	return false, nil
}

func updateSelf() {
	resp, err := http.Get("https://s3.amazonaws.com/progrium-flynn/flynn-grid/dev/grid-cli_" + runtime.GOOS + "_" + runtime.GOARCH + ".tgz")
	assert(err)
	defer resp.Body.Close()
	z, err := gzip.NewReader(resp.Body)
	assert(err)
	defer z.Close()
	t := tar.NewReader(z)
	hdr, err := t.Next()
	assert(err)
	if hdr.Name != "grid" {
		log.Fatal("grid binary not found in tarball")
	}
	selfpath, err := osext.Executable()
	assert(err)
	info, err := os.Stat(selfpath)
	assert(err)
	assert(os.Rename(selfpath, selfpath+".old"))
	f, err := os.OpenFile(selfpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		assert(os.Rename(selfpath+".old", selfpath))
		assert(err)
	}
	defer f.Close()
	_, err = io.Copy(f, t)
	if err != nil {
		assert(os.Rename(selfpath+".old", selfpath))
		assert(err)
	}
	assert(os.Remove(selfpath + ".old"))
	fmt.Println("Updated.")
}
