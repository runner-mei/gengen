package main

import (
	"errors"
	"flag"
	"path/filepath"
	"text/template"

	"github.com/three-plus-three/gengen/types"
)

// GenerateJSCommand - 生成视图
type GenerateJSCommand struct {
	baseCommand
}

// Flags - 申明参数
func (cmd *GenerateJSCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	return cmd.baseCommand.Flags(fs)
}

// Run - 生成代码
func (cmd *GenerateJSCommand) Run(args []string) error {
	return cmd.run(args, cmd.genrateJS)
}

func (cmd *GenerateJSCommand) genrateJS(cls *types.ClassSpec) error {
	ctlName := Pluralize(cls.Name)
	params := map[string]interface{}{"namespace": cmd.ns,
		"controllerName": ctlName,
		"modelName":      ctlName,
		"class":          cls}
	funcs := template.FuncMap{}

	err := cmd.executeTempate(cmd.override, []string{"views/js"}, funcs, params,
		filepath.Join(cmd.output, Underscore(Pluralize(cls.Name)), Underscore(Pluralize(cls.Name))+".js"))
	if err != nil {
		return errors.New("gen views/js: " + err.Error())
	}
	return nil
}
