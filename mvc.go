package main

import (
	"flag"
	"path/filepath"
)

// GenerateMVCCommand - 生成代码
type GenerateMVCCommand struct {
	baseCommand
	controller  string
	projectPath string
	layouts     string
	customPath  string
}

// Flags - 申明参数
func (cmd *GenerateMVCCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.controller, "controller", "", "the base controller name")
	fs.StringVar(&cmd.projectPath, "projectPath", "", "the project path")
	fs.StringVar(&cmd.layouts, "layouts", "", "")
	fs.StringVar(&cmd.customPath, "customPath", "", "")
	return cmd.baseCommand.Flags(fs)
}

// Run - 生成数据库模型代码
func (cmd *GenerateMVCCommand) Run(args []string) error {
	var st GenerateStructCommand
	var views GenerateViewCommand
	var js GenerateJSCommand
	var ctl GenerateControllerCommand
	var ut GenerateUnitTestCommand
	ctl.ns = "models"
	st.CopyFrom(&cmd.baseCommand)
	st.output = filepath.Join(cmd.output, "app", "models")
	views.CopyFrom(&cmd.baseCommand)
	views.layouts = cmd.layouts
	views.customPath = cmd.customPath
	views.output = filepath.Join(cmd.output, "app", "views")
	js.CopyFrom(&cmd.baseCommand)
	js.output = filepath.Join(cmd.output, "public", "js")
	ctl.CopyFrom(&cmd.baseCommand)
	ctl.ns = "controllers"
	ctl.controller = cmd.controller
	ctl.projectPath = cmd.projectPath
	ctl.output = filepath.Join(cmd.output, "app", "controllers")
	ut.CopyFrom(&cmd.baseCommand)
	ut.ns = "tests"
	// ut.controller = cmd.controller
	ut.projectPath = cmd.projectPath
	ut.output = cmd.output

	if err := st.Run(args); err != nil {
		return err
	}

	if err := views.Run(args); err != nil {
		return err
	}

	if err := js.Run(args); err != nil {
		return err
	}

	if err := ctl.Run(args); err != nil {
		return err
	}

	if err := ut.Run(args); err != nil {
		return err
	}
	return nil
}
