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
