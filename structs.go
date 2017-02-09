package main

import (
	"cn/com/hengwei/commons/types"
	"path/filepath"
	"text/template"
)

// GenerateModelsCommand - 生成数据库模型代码
type GenerateStructCommand struct {
	baseCommand
}

// Run - 生成数据库模型代码
func (cmd *GenerateStructCommand) Run(args []string) error {
	return cmd.run(args, cmd.generateStruct)
}

func (cmd *GenerateStructCommand) generateStruct(cls *types.TableDefinition) error {
	funcs := template.FuncMap{"omitempty": func(t *types.PropertyDefinition) bool {
		return !t.IsRequired
	}}
	params := map[string]interface{}{"namespace": cmd.ns,
		"class": cls}

	return cmd.executeTempate(cmd.override, "struct", funcs, params,
		filepath.Join(cmd.output, cls.UnderscoreName+".go"))
}
