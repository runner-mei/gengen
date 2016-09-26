package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
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

func (cmd *versionCommand) Run(args []string) {
	fmt.Println(filepath.Base(os.Args[0]), "1.0")
}

type embedeCommand struct {
	//file string
}

func (cmd *embedeCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	// define subcommand's flags
	//fs.StringVar(&cmd.file, "file", "", "provides")
	return fs
}

func (cmd *embedeCommand) Run(args []string) {
	if len(args) != 2 {
		fmt.Println("argument error.")
		return
	}
	out, e := os.OpenFile(args[0], os.O_TRUNC|os.O_RDWR|os.O_CREATE, 0)
	if nil != e {
		fmt.Println(e)
		return
	}
	defer out.Close()
	in, e := os.OpenFile(args[1], os.O_RDONLY, 0)
	if nil != e {
		fmt.Println(e)
		return
	}
	defer in.Close()

	out.WriteString(`package main

var embede_text = `)
	out.WriteString("`")
	if _, e = io.Copy(out, in); nil != e {
		fmt.Println(e)
		return
	}
	out.WriteString("`")
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
	dbBase
	root     string
	override bool
}

func (cmd *generateCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	cmd.initFlags(fs)
	flag.StringVar(&cmd.root, "root", "", "the root directory")
	flag.BoolVar(&cmd.override, "override", false, "")
	return fs
}

func (cmd *generateCommand) init() error {
	if "" == cmd.root {
		for _, s := range []string{"conf/routes", "../conf/routes", "../../conf/routes", "../../conf/routes"} {
			if st, e := os.Stat(s); nil == e && nil != st && !st.IsDir() {
				cmd.root, _ = filepath.Abs(filepath.Join(s, "..", ".."))
				break
			}
		}

		if "" == cmd.root {
			return errors.New("root directory isn't found")
		}
	}
	return nil
}

func (cmd *generateCommand) Run(args []string) {
	if len(args) == 0 {
		fmt.Println("table name is missing.")
		return
	}

	if e := cmd.init(); nil != e {
		fmt.Println(e)
		return
	}

	var controller = GenerateControllerCommand{dbBase: cmd.dbBase, root: cmd.root}
	var views = GenerateViewCommand{dbBase: cmd.dbBase, root: cmd.root}

	controller.Run(args)
	views.Run(args)
}
