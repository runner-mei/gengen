package main

import (
	"flag"
	"path/filepath"
	"text/template"

	"github.com/three-plus-three/gengen/types"
)

// GenerateControllerCommand - 生成控制器
type GenerateControllerCommand struct {
	baseCommand
	controller  string
	projectPath string
}

// Flags - 申明参数
func (cmd *GenerateControllerCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.controller, "controller", "", "the base controller name")
	fs.StringVar(&cmd.projectPath, "projectPath", "", "the project path")
	return cmd.baseCommand.Flags(fs)
}

// Run - 生成代码
func (cmd *GenerateControllerCommand) Run(args []string) error {
	return cmd.run(args, func(cls *types.ClassSpec) error {
		funcs := template.FuncMap{"displayForBelongsTo": func(f types.FieldSpec) string {
			ann := f.Annotations["display"]
			if s, ok := ann.(string); ok {
				return Goify(s, true)
			}

			fields := ReferenceFields(f)
			if len(fields) >= 1 {
				return Goify(fields[0].Name, true)
			}

			return "Name"
		}}

		params := map[string]interface{}{"namespace": cmd.ns,
			"baseController": cmd.controller,
			"projectPath":    cmd.projectPath,
			"controllerName": Pluralize(cls.Name),
			"modelName":      Pluralize(cls.Name),
			"class":          cls}

		return cmd.executeTempate(cmd.override, []string{"ns", "controller"}, funcs, params,
			filepath.Join(cmd.output, Underscore(Pluralize(cls.Name))+".go"))
	})
}
