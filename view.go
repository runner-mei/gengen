package main

import (
	"cn/com/hengwei/commons/types"
	"os"
	"path/filepath"
	"text/template"
)

// GenerateViewCommand - 生成视图
type GenerateViewCommand struct {
	baseCommand
}

// Run - 生成代码
func (cmd *GenerateViewCommand) Run(args []string) error {
	return cmd.run(args, cmd.genrateView)
}

func (cmd *GenerateViewCommand) genrateView(cls *types.TableDefinition) error {
	params := map[string]interface{}{"namespace": cmd.ns,
		"controllerName": types.Pluralize(cls.Name),
		"class":          cls}
	funcs := template.FuncMap{}

	err := cmd.executeTempate(cmd.override, "views/index", funcs, params, filepath.Join(cmd.output, "index.html"))
	if err != nil {
		return err
	}

	err = cmd.executeTempate(cmd.override, "views/edit", funcs, params, filepath.Join(cmd.output, "edit.html"))
	if err != nil {
		os.Remove(filepath.Join(cmd.output, "index.html"))
		return err
	}

	err = cmd.executeTempate(cmd.override, "views/fields", funcs, params, filepath.Join(cmd.output, "edit_fields.html"))
	if err != nil {
		os.Remove(filepath.Join(cmd.output, "index.html"))
		os.Remove(filepath.Join(cmd.output, "edit.html"))
		return err
	}

	err = cmd.executeTempate(cmd.override, "views/new", funcs, params, filepath.Join(cmd.output, "new.html"))
	if err != nil {
		os.Remove(filepath.Join(cmd.output, "index.html"))
		os.Remove(filepath.Join(cmd.output, "edit.html"))
		os.Remove(filepath.Join(cmd.output, "edit_fields.html"))
		return err
	}

	err = cmd.executeTempate(cmd.override, "views/quick", funcs, params, filepath.Join(cmd.output, "quick-bar.html"))
	if err != nil {
		os.Remove(filepath.Join(cmd.output, "index.html"))
		os.Remove(filepath.Join(cmd.output, "edit.html"))
		os.Remove(filepath.Join(cmd.output, "edit_fields.html"))
		os.Remove(filepath.Join(cmd.output, "new.html"))
		return err
	}
	return nil
}
