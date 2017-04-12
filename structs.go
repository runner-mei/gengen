package main

import (
	"path/filepath"
	"text/template"

	"github.com/runner-mei/gengen/types"
)

// GenerateStructCommand - 生成数据库模型代码
type GenerateStructCommand struct {
	baseCommand
}

// Run - 生成数据库模型代码
func (cmd *GenerateStructCommand) Run(args []string) error {
	return cmd.run(args, cmd.generateStruct)
}

func (cmd *GenerateStructCommand) generateStruct(cls *types.ClassSpec) error {
	funcs := template.FuncMap{
		"omitempty": func(t types.FieldSpec) bool {
			return !t.IsRequired
		}}
	params := map[string]interface{}{"namespace": cmd.ns,
		"class": cls}

	return cmd.executeTempate(cmd.override, []string{"ns", "struct"}, funcs, params,
		filepath.Join(cmd.output, Underscore(Pluralize(cls.Name))+".go"))
}
