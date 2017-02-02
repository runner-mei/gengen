package main

import (
	"cn/com/hengwei/commons/types"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"
)

// GenerateModelsCommand - 生成数据库模型代码
type GenerateStructCommand struct {
	ns     string
	root   string
	output string
	theme  string
}

// Flags - 申明参数
func (cmd *GenerateStructCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.ns, "namespace", "models", "the namespace name")
	fs.StringVar(&cmd.root, "root", "", "the input target")
	fs.StringVar(&cmd.output, "output", "", "the output target")
	fs.StringVar(&cmd.theme, "theme", "", "the theme target")
	return fs
}

func (cmd *GenerateStructCommand) loadFile(nm string) ([]byte, error) {
	if cmd.theme != "" {
		file := filepath.Join(cmd.root, cmd.theme, nm+".tpl.go")
		bs, e := ioutil.ReadFile(file)
		if e == nil {
			return bs, nil
		}
		if !os.IsNotExist(e) {
			return nil, errors.New("load template fail, " + e.Error())
		}
	}

	file := filepath.Join(cmd.root, "default", nm+".tpl.go")
	bs, e := ioutil.ReadFile(file)
	if e == nil {
		return bs, nil
	}
	if !os.IsNotExist(e) {
		return nil, errors.New("load template fail, " + e.Error())
	}

	return []byte(structText), nil
}

// Run - 生成数据库模型代码
func (cmd *GenerateStructCommand) Run(args []string) error {
	// if e := cmd.init(); e != nil {
	// 	return e
	// }

	tables, e := cmd.loadTables()
	if nil != e {
		return e
	}
	if len(args) > 0 {
		for _, name := range args {
			log.Println("[GEN] ", name)
			table := tables.Find(name)
			if table == nil {
				table = tables.FindByUnderscoreName(name)
				if table == nil {
					table = tables.FindByTableName(name)
				}
			}

			if table == nil {
				log.Println("[FAIL]", name)
				return errors.New("'" + name + "' isn't found")
			}

			var out io.Writer = os.Stdout
			// switch strings.ToLower(cmd.file) {
			// case "stdout":
			// 	out = os.Stdout
			// case "stderr":
			// 	out = os.Stderr
			// case "":
			// 	out = os.Stderr
			// default:
			// 	f, e := os.Create(filepath.Join(cmd.output, table.UnderscoreName+"_gen.go"))
			// 	if e != nil {
			// 		return e
			// 	}
			// 	defer f.Close()
			//
			// 	out = f
			// }

			if e := cmd.genrateStruct(out, table); nil != e {
				return e
			}
		}
	} else {
		for _, table := range tables.All() {
			log.Println("[GEN] ", table.Name)

			f, e := os.Create(filepath.Join(cmd.output, table.UnderscoreName+"_gen.go"))
			if e != nil {
				return e
			}
			defer f.Close()

			if e := cmd.genrateStruct(f, table); nil != e {
				return e
			}
		}
	}
	return nil
}

func (cmd *GenerateStructCommand) genrateStruct(out io.Writer, cls *types.TableDefinition) error {
	funcs := template.FuncMap{"goify": Goify,
		"gotype":      GoTypename,
		"underscore":  Underscore,
		"tableize":    Tableize,
		"singularize": Singularize,
		"omitempty": func(t *types.PropertyDefinition) bool {
			return !t.IsRequired
		}}

	bs, e := cmd.loadFile("struct")
	if e != nil {
		return e
	}

	t, e := template.New("").Funcs(funcs).Parse(string(bs))
	if e != nil {
		return e
	}
	return t.Execute(out, map[string]interface{}{"namespace": cmd.ns, "class": cls})
}

func (cmd *GenerateStructCommand) loadTables() (*types.TableDefinitions, error) {
	files, err := filepath.Glob(filepath.Join(cmd.root, "*"))
	if err != nil {
		return nil, err
	}

	return types.LoadFiles(files)
}
