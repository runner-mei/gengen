package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"path/filepath"
)

// GenerateMVCCommand - 生成代码
type GenerateMVCCommand struct {
	baseCommand
	controller  string
	projectPath string
	layouts     string
	customPath  string
	viewTag     string
}

// Flags - 申明参数
func (cmd *GenerateMVCCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.controller, "controller", "", "the base controller name")
	fs.StringVar(&cmd.projectPath, "projectPath", "", "the project path")
	fs.StringVar(&cmd.layouts, "layouts", "", "")
	fs.StringVar(&cmd.customPath, "customPath", "", "")
	fs.StringVar(&cmd.viewTag, "view_tag", "", "")
	return cmd.baseCommand.Flags(fs)
}

// Run - 生成数据库模型代码
func (cmd *GenerateMVCCommand) Run(args []string) error {
	go http.ListenAndServe(":", nil)

	var st GenerateStructCommand
	var views GenerateViewCommand
	var js GenerateJSCommand
	var ctl GenerateControllerCommand
	var ut GenerateUnitTestCommand

	st.ns = "models"
	st.theme = cmd.theme
	st.CopyFrom(&cmd.baseCommand)
	st.output = filepath.Join(cmd.output, "app", "models")
	views.CopyFrom(&cmd.baseCommand)
	views.ns = "views"
	views.theme = cmd.theme
	views.layouts = cmd.layouts
	views.customPath = cmd.customPath
	views.viewTag = cmd.viewTag
	views.output = filepath.Join(cmd.output, "app", "views")
	js.CopyFrom(&cmd.baseCommand)
	js.ns = "js"
	js.theme = cmd.theme
	js.output = filepath.Join(cmd.output, "public", "js")
	ctl.CopyFrom(&cmd.baseCommand)
	ctl.ns = "controllers"
	ctl.theme = cmd.theme
	ctl.controller = cmd.controller
	ctl.projectPath = cmd.projectPath
	ctl.output = filepath.Join(cmd.output, "app", "controllers")
	ut.CopyFrom(&cmd.baseCommand)
	ut.ns = "tests"
	ut.theme = cmd.theme
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
