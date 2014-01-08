package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"sort"
	"text/tabwriter"
)

var cmdVersion = &Command{
	Run:   runVersion,
	Usage: "version",
	Short: "show grid version",
	Long:  `Version shows the grid utility version string.`,
}

func runVersion(cmd *Command, args []string) {
	fmt.Println(Version)
}

var cmdHelp = &Command{
	Usage: "help [<topic>]",
	Long:  `Help shows usage for a command or other topic.`,
}

var helpCommands = &Command{
	Usage: "commands",
	Short: "list all commands with usage",
	Long:  "(not displayed; see special case in runHelp)",
}

func init() {
	cmdHelp.Run = runHelp // break init loop
}

func runHelp(cmd *Command, args []string) {
	if len(args) == 0 {
		printUsage()
		return // not os.Exit(2); success
	}
	if len(args) != 1 {
		log.Fatal("too many arguments")
	}
	switch args[0] {
	case helpCommands.Name():
		printAllUsage()
		return
	}

	for _, cmd := range commands {
		if cmd.Name() == args[0] {
			cmd.printUsage()
			return
		}
	}

	log.Printf("Unknown help topic: %q. Run 'grid help'.\n", args[0])
	os.Exit(2)
}

func maxStrLen(strs []string) (strlen int) {
	for i := range strs {
		if len(strs[i]) > strlen {
			strlen = len(strs[i])
		}
	}
	return
}

var usageTemplate = template.Must(template.New("usage").Parse(`
Usage: grid <command> [options] [arguments]


Commands:
{{range .Commands}}{{if .Runnable}}{{if .List}}
    {{.Name | printf (print "%-" $.MaxRunListName "s")}}  {{.Short}}{{end}}{{end}}{{end}}

Run 'grid help [command]' for details.


Additional help topics:
{{range .Commands}}{{if not .Runnable}}
    {{.Name | printf "%-8s"}}  {{.Short}}{{end}}{{end}}

`[1:]))

func printUsage() {
	var runListNames []string
	for i := range commands {
		if commands[i].Runnable() && commands[i].List() {
			runListNames = append(runListNames, commands[i].Name())
		}
	}

	usageTemplate.Execute(os.Stderr, struct {
		Commands       []*Command
		MaxRunListName int
	}{
		commands,
		maxStrLen(runListNames),
	})
}

func printAllUsage() {
	w := tabwriter.NewWriter(os.Stdout, 1, 2, 2, ' ', 0)
	defer w.Flush()
	cl := commandList(commands)
	sort.Sort(cl)
	for i := range cl {
		if cl[i].Runnable() {
			listRec(w, "grid "+cl[i].FullUsage(), "# "+cl[i].Short)
		}
	}
}

func usage() {
	printUsage()
	os.Exit(2)
}

type commandList []*Command

func (cl commandList) Len() int           { return len(cl) }
func (cl commandList) Swap(i, j int)      { cl[i], cl[j] = cl[j], cl[i] }
func (cl commandList) Less(i, j int) bool { return cl[i].Name() < cl[j].Name() }

type commandMap map[string]commandList
