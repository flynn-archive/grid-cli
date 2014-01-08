package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func targetPath() string {
	return filepath.Join(homePath(), ".thegrid")
}

func getTarget() string {
	if _, err := os.Stat(targetPath()); os.IsNotExist(err) {
		return "127.0.0.1"
	}
	data, _ := ioutil.ReadFile(targetPath())
	return string(data)
}

func setTarget(addr string) error {
	return ioutil.WriteFile(targetPath(), []byte(addr), 0644)
}

func homePath() string {
	u, err := user.Current()
	if err != nil {
		panic("couldn't determine user: " + err.Error())
	}
	return u.HomeDir
}

type Command struct {
	// args does not include the command name
	Run  func(cmd *Command, args []string)
	Flag flag.FlagSet

	Usage string // first word is the command name
	Short string // `grid help` output
	Long  string // `grid help cmd` output
}

func (c *Command) printUsage() {
	if c.Runnable() {
		fmt.Printf("Usage: grid %s\n\n", c.FullUsage())
	}
	fmt.Println(strings.Trim(c.Long, "\n"))
}

func (c *Command) FullUsage() string {
	return c.Usage
}

func (c *Command) Name() string {
	name := c.Usage
	i := strings.Index(name, " ")
	if i >= 0 {
		name = name[:i]
	}
	return name
}

func (c *Command) Runnable() bool {
	return c.Run != nil
}

func (c *Command) List() bool {
	return c.Short != ""
}

// Running `grid help` will list commands in this order.
var commands = []*Command{
	cmdJobs,
	cmdTarget,
	cmdVersion,
	cmdHelp,

	helpAbout,
}

func main() {
	log.SetFlags(0)

	// make sure command is specified, disallow global args
	args := os.Args[1:]
	if len(args) < 1 || strings.IndexRune(args[0], '-') == 0 {
		usage()
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] && cmd.Run != nil {
			cmd.Flag.Usage = func() {
				cmd.printUsage()
			}
			if err := cmd.Flag.Parse(args[1:]); err != nil {
				os.Exit(2)
			}
			cmd.Run(cmd, cmd.Flag.Args())
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown command: %s\n", args[0])
	usage()
}

func listRec(w io.Writer, a ...interface{}) {
	for i, x := range a {
		fmt.Fprint(w, x)
		if i+1 < len(a) {
			w.Write([]byte{'\t'})
		} else {
			w.Write([]byte{'\n'})
		}
	}
}

func assert(err error) error {
	if err != nil {
		log.Fatal(err)
	}
	return err
}
