package main

import (
	"flag"
	"path/filepath"
	"text/template"

	"github.com/three-plus-three/gengen/types"
)

// GenerateModelsCommand - 生成数据库模型代码
type GenerateTestCommand struct {
	baseCommand
	projectPath string
}

// Flags - 申明参数
func (cmd *GenerateTestCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.projectPath, "projectPath", "", "the project path")
	return cmd.baseCommand.Flags(fs)
}

// Run - 生成数据库模型代码
func (cmd *GenerateTestCommand) Run(args []string) error {
	return cmd.runAll(args, cmd.generateStructs)
}

func (cmd *GenerateTestCommand) generateStructs(tables []*types.ClassSpec) error {
	funcs := template.FuncMap{
		"omitempty": func(t types.FieldSpec) bool {
			return !t.IsRequired
		},
		"tableName": getTableName}

	params := map[string]interface{}{"namespace": cmd.ns,
		"classes":     tables,
		"projectPath": cmd.projectPath}

	return cmd.executeTempate(cmd.override, []string{"tests/test_base"}, funcs, params,
		filepath.Join(cmd.output, "tests", "test_base.go"))
}
