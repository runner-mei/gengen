package main

import (
	"cn/com/hengwei/commons/types"
	"path/filepath"
	"text/template"
)

// GenerateControllerCommand - 生成控制器
type GenerateControllerCommand struct {
	baseCommand
}

// Run - 生成代码
func (cmd *GenerateControllerCommand) Run(args []string) error {
	return cmd.run(args, func(cls *types.TableDefinition) error {
		funcs := template.FuncMap{}

		params := map[string]interface{}{"namespace": cmd.ns,
			"controllerName": types.Pluralize(cls.Name),
			"class":          cls}

		return cmd.executeTempate(cmd.override, "controller", funcs, params,
			filepath.Join(cmd.output, cls.UnderscoreName+".go"))
	})
}
