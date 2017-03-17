package main

import (
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/runner-mei/gengen/types"
)

// GenerateModelsCommand - 生成数据库模型代码
type baseCommand struct {
	ns       string
	root     string
	output   string
	theme    string
	override bool
}

func (cmd *baseCommand) CopyFrom(b *baseCommand) {
	cmd.ns = b.ns
	cmd.root = b.root
	cmd.output = b.output
	cmd.theme = b.theme
	cmd.override = b.override
}

// Flags - 申明参数
func (cmd *baseCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.ns, "namespace", "models", "the namespace name")
	fs.StringVar(&cmd.root, "root", "", "the input target")
	fs.StringVar(&cmd.output, "output", "", "the output target")
	fs.StringVar(&cmd.theme, "theme", "", "the theme target")
	fs.BoolVar(&cmd.override, "override", false, "")
	return fs
}

func (cmd *baseCommand) loadTables() ([]*types.ClassSpec, error) {
	files, err := filepath.Glob(filepath.Join(cmd.root, "*"))
	if err != nil {
		return nil, errors.New("search root directory fail, " + err.Error())
	}

	return types.LoadYAMLFiles(files)
}

func (cmd *baseCommand) loadFile(nm string) ([]byte, error) {
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

	return textDefault(nm), nil
}

func (cmd *baseCommand) newTemplate(name string, funcs template.FuncMap) (*template.Template, error) {
	locals := template.FuncMap{"goify": Goify,
		"gotype":            GoTypename,
		"underscore":        types.Underscore,
		"tableize":          types.Tableize,
		"singularize":       types.Singularize,
		"pluralize":         types.Pluralize,
		"camelizeDownFirst": types.CamelizeDownFirst,
		"omitempty": func(t *types.FieldSpec) bool {
			return !t.IsRequired
		}}

	bs, e := cmd.loadFile(name)
	if e != nil {
		return nil, e
	}

	return template.New("").Delims("[[", "]]").Funcs(funcs).Funcs(locals).Parse(string(bs))
}

// Run - 生成数据库模型代码
func (cmd *baseCommand) runAll(args []string, cb func(tables []*types.ClassSpec) error) error {
	// if e := cmd.init(); e != nil {
	//  return e
	// }

	if cmd.output != "" {
		if st, err := os.Stat(cmd.output); err != nil {
			if os.IsNotExist(err) {
				return err
			}

			if err := os.MkdirAll(cmd.output, 0); err != nil {
				return err
			}
		} else if !st.IsDir() {
			return errors.New(("'" + cmd.output + "' isn't directory."))
		}
	}

	tables, e := cmd.loadTables()
	if nil != e {
		return e
	}
	return cb(tables)
}

// Run - 生成数据库模型代码
func (cmd *baseCommand) run(args []string, cb func(table *types.ClassSpec) error) error {
	// if e := cmd.init(); e != nil {
	//  return e
	// }

	if cmd.output != "" {
		if st, err := os.Stat(cmd.output); err != nil {
			if !os.IsNotExist(err) {
				return err
			}

			if err := os.MkdirAll(cmd.output, 0); err != nil {
				return err
			}
		} else if !st.IsDir() {
			return errors.New(("'" + cmd.output + "' isn't directory."))
		}
	}

	tables, e := cmd.loadTables()
	if nil != e {
		return e
	}
	if len(args) > 0 {
		for _, name := range args {
			log.Println("[GEN] ", name)
			var table *types.ClassSpec
			for _, cs := range tables {
				if cs.Name == name {
					table = cs
					break
				}
			}

			if table == nil {
				log.Println("[FAIL]", name)
				return errors.New("'" + name + "' isn't found")
			}

			// var out io.Writer = os.Stdout
			// switch strings.ToLower(cmd.file) {
			// case "stdout":
			//  out = os.Stdout
			// case "stderr":
			//  out = os.Stderr
			// case "":
			//  out = os.Stderr
			// default:
			//  f, e := os.Create(filepath.Join(cmd.output, table.UnderscoreName+"_gen.go"))
			//  if e != nil {
			//    return e
			//  }
			//  defer f.Close()
			//
			//  out = f
			// }

			if e := cb(table); nil != e {
				return e
			}
		}
	} else {
		for _, table := range tables {
			log.Println("[GEN] ", table.Name)

			// f, e := os.Create(filepath.Join(cmd.output, table.UnderscoreName+"_gen.go"))
			// if e != nil {
			//  return e
			// }
			// defer f.Close()

			if e := cb(table); nil != e {
				return e
			}
		}
	}
	return nil
}

func (cmd *baseCommand) executeTempate(override bool, names []string, funcs template.FuncMap, params interface{}, fname string) error {
	var out *os.File
	var err error

	dirname := filepath.Dir(fname)
	if dirname != "" {
		if err = os.MkdirAll(dirname, 0666); err != nil {
			if !os.IsExist(err) {
				return err
			}
		}
	}

	if !override {
		out, err = os.OpenFile(fname, os.O_CREATE|os.O_EXCL, 0666)
	} else {
		out, err = os.OpenFile(fname, os.O_CREATE|os.O_TRUNC, 0666)
	}
	if nil != err {
		return err
	}
	defer out.Close()

	for _, name := range names {
		tpl, err := cmd.newTemplate(name, funcs)
		if nil != err {
			return err
		}

		if err := tpl.Execute(out, params); err != nil {
			out.Close()
			os.Remove(fname)
			return err
		}
	}
	return nil
}
