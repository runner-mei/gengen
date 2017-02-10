package main

import (
	"cn/com/hengwei/commons/types"
	"flag"
	"html/template"
	"path/filepath"
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
	return cmd.run(args, func(cls *types.TableDefinition) error {
		funcs := template.FuncMap{}

		if cmd.handlerPrefix == "" {
			cmd.handlerPrefix = cls.Name
		}

		params := map[string]interface{}{"namespace": cmd.ns,
			"handlerPrefix": cmd.handlerPrefix,
			"class":         cls}

		return cmd.executeTempate(cmd.override, []string{"ns", "struct", "handler"}, funcs, params,
			filepath.Join(cmd.output, cls.UnderscoreName+".go"))
	})
}
