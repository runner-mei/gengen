package main

import (
	"flag"
	"path/filepath"
	"text/template"

	"github.com/runner-mei/gengen/types"
)

// HandlerCommand - 生成控制器
type HandlerCommand struct {
	baseCommand
	handlerPrefix string
}

// Flags - 申明参数
func (cmd *HandlerCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.handlerPrefix, "handlerPrefix", "", "")
	return cmd.baseCommand.Flags(fs)
}

// Run - 生成代码
func (cmd *HandlerCommand) Run(args []string) error {
	return cmd.run(args, func(cls *types.ClassSpec) error {
		funcs := template.FuncMap{}

		if cmd.handlerPrefix == "" {
			cmd.handlerPrefix = cls.Name
		}

		params := map[string]interface{}{"namespace": cmd.ns,
			"handlerPrefix": cmd.handlerPrefix,
			"class":         cls}

		return cmd.executeTempate(cmd.override, []string{"ns", "struct", "handler"}, funcs, params,
			filepath.Join(cmd.output, Underscore(cls.Name)+".go"))
	})
}
