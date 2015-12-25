package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rakyll/command"
)

type versionCommand struct {
	flagVerbose *bool
}

func (cmd *versionCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	// define subcommand's flags
	cmd.flagVerbose = fs.Bool("v", false, "provides verbose output")
	return fs
}

func (cmd *versionCommand) Run(args []string) {
	fmt.Println(filepath.Base(os.Args[0]), "1.0")
}

func init() {
	// register version as a subcommand
	command.On("version", "prints the version", &versionCommand{}, nil)
}

func main() {
	flag.Usage = command.Usage
	command.Parse()
	command.Run()
}
