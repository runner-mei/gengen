package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/runner-mei/command"
)

type versionCommand struct {
	flagVerbose *bool
}

func (cmd *versionCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	// define subcommand's flags
	cmd.flagVerbose = fs.Bool("v", false, "provides verbose output")
	return fs
}

func (cmd *versionCommand) Run(args []string) error {
	fmt.Println(filepath.Base(os.Args[0]), "1.0")
	return nil
}

type embedeCommand struct {
	//file string
}

func (cmd *embedeCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	// define subcommand's flags
	//fs.StringVar(&cmd.file, "file", "", "provides")
	return fs
}

func (cmd *embedeCommand) Run(args []string) error {
	if len(args) != 1 {
		return errors.New("argument error")
	}
	out, e := os.OpenFile(args[0], os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0)
	if nil != e {
		return e
	}
	defer out.Close()
	out.WriteString(`package main
`)

	for _, t := range [][2]string{{"embede_text", "base.go"},
		{"structText", "tpl/struct.go"}} {
		if e := cmd.copyFile(out, t[0], t[1]); e != nil {
			return e
		}
	}
	return nil
}

func (cmd *embedeCommand) copyFile(out io.Writer, name, file string) error {
	bs, e := ioutil.ReadFile(file)
	if nil != e {
		return e
	}

	io.WriteString(out, `

var `+name+` = `)
	io.WriteString(out, "`")

	out.Write(bytes.Replace(bs, []byte("`"), []byte("` + \"`\" + `"), -1))

	_, e = io.WriteString(out, "`")
	return e
}

func init() {
	// register version as a subcommand
	command.On("version", "prints the version", &versionCommand{}, nil)
	command.On("embede", "", &embedeCommand{}, nil)
	command.On("generate", "从数据库的表模型生成控制器和 views 代码", &generateCommand{}, nil)
	command.On("models", "从数据库的表模型生成 models 代码", &GenerateModelsCommand{}, nil)
	command.On("controller", "从数据库的表模型生成控制器代码", &GenerateControllerCommand{}, nil)
	command.On("views", "从数据库的表模型生成 Views 代码", &GenerateViewCommand{}, nil)
}

func main() {
	flag.Usage = command.Usage
	command.Parse()
	command.Run()
}

type generateCommand struct {
	generateBase
}

func (cmd *generateCommand) Run(args []string) error {
	if len(args) == 0 {
		return errors.New("table name is missing.")
	}

	if e := cmd.init(); nil != e {
		return e
	}

	var controller = GenerateControllerCommand{generateBase: cmd.generateBase}
	var views = GenerateViewCommand{generateBase: cmd.generateBase}

	controller.Run(args)
	views.Run(args)
	return nil
}
