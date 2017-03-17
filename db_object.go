package main

import (
	"path/filepath"
	"text/template"

	"github.com/runner-mei/gengen/types"
)

// GenerateModelsCommand - 生成数据库模型代码
type GenerateDBObjectCommand struct {
	baseCommand
}

// Run - 生成数据库模型代码
func (cmd *GenerateDBObjectCommand) Run(args []string) error {
	return cmd.runAll(args, cmd.generateStructs)
}

func (cmd *GenerateDBObjectCommand) generateStructs(tables []*types.ClassSpec) error {
	funcs := template.FuncMap{
		"omitempty": func(t types.FieldSpec) bool {
			return !t.IsRequired
		},
		"tableName": func(t types.ClassSpec) string {
			if t.Table != "" {
				return t.Table
			}
			return Tableize(t.Name)
		}}

	params := map[string]interface{}{"namespace": cmd.ns,
		"classes": tables}

	return cmd.executeTempate(cmd.override, []string{"ns", "db"}, funcs, params,
		filepath.Join(cmd.output, "db.go"))
}
