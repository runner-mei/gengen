package main

// // GenerateModelsCommand - 生成数据库模型代码
// type GenerateModelsCommand struct {
// 	GenerateStructCommand
// }

// // Run - 生成数据库模型代码
// func (cmd *GenerateModelsCommand) Run(args []string) error {
// 	return cmd.run(args, cmd.generateModel)
// }

// func (cmd *GenerateModelsCommand) generateModel(cls *types.TableDefinition) error {
// 	funcs := template.FuncMap{"omitempty": func(t *types.PropertyDefinition) bool {
// 		return !t.IsRequired
// 	}}
// 	params := map[string]interface{}{"namespace": cmd.ns,
// 		"class": cls}

// 	return cmd.executeTempate(cmd.override, []string{"struct"}, funcs, params,
// 		filepath.Join(cmd.output, cls.UnderscoreName+".go"))
// }
