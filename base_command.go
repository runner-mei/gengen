package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/three-plus-three/gengen/types"
)

// GenerateModelsCommand - 生成数据库模型代码
type baseCommand struct {
	ns       string
	root     string
	output   string
	theme    string
	override bool
	funcs    template.FuncMap
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
	locals := template.FuncMap{
		"set": func(ctx map[string]interface{}, name string, value interface{}) string {
			ctx[name] = value
			return ""
		},
		"goify":             Goify,
		"gotype":            GoTypename,
		"underscore":        types.Underscore,
		"tableize":          types.Tableize,
		"singularize":       types.Singularize,
		"pluralize":         types.Pluralize,
		"camelizeDownFirst": types.CamelizeDownFirst,
		"toFormat":          toFormatFunc,
		"omitempty": func(t *types.FieldSpec) bool {
			return !t.IsRequired
		},
		"editDisabled": func(f interface{}) bool {
			return HasFeature(f, "editDisabled")
		},
		"newDisabled": func(f interface{}) bool {
			return HasFeature(f, "newDisabled")
		},
		"deleteDisabled": func(f interface{}) bool {
			return HasFeature(f, "deleteDisabled")
		},
		"valueInAnnotations": ValueInAnnotations,
		"hasFeature":         HasFeature,
		"hasAnyFeatures": func(f interface{}, names ...string) bool {
			for _, nm := range names {
				if HasFeature(f, nm) {
					return true
				}
			}
			return false
		},
		"hasAllFeatures": func(f interface{}, names ...string) bool {
			for _, nm := range names {
				if !HasFeature(f, nm) {
					return false
				}
			}
			return true
		},
		"fieldExists": func(cls *types.ClassSpec, fieldName string) bool {
			for _, field := range cls.Fields {
				if field.Name == fieldName {
					return true
				}
			}
			return false
		},
		"field": func(cls *types.ClassSpec, fieldName string) types.FieldSpec {
			for _, field := range cls.Fields {
				if field.Name == fieldName {
					return field
				}
			}
			panic(errors.New("field '" + fieldName + "' isn't exists in the " + cls.Name))
		},
		"hasEnumerations": func(f types.FieldSpec) bool {
			if f.Restrictions == nil {
				return false
			}
			return len(f.Restrictions.Enumerations) > 0
		},
		"isBelongsTo": func(cls *types.ClassSpec, f types.FieldSpec) bool {
			for _, belongsTo := range cls.BelongsTo {
				if belongsTo.Name == f.Name {
					return true
				}
			}
			return false
		},
		"belongsTo": func(cls *types.ClassSpec, f types.FieldSpec) *types.BelongsTo {
			for _, belongsTo := range cls.BelongsTo {
				if belongsTo.Name == f.Name {
					return &belongsTo
				}
			}
			return nil
		}}

	if len(cmd.funcs) != 0 {
		for k, v := range cmd.funcs {
			funcs[k] = v
		}
	}

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
	if cmd.funcs == nil {
		cmd.funcs = template.FuncMap{}
	}
	cmd.funcs["class"] = func(name string) *types.ClassSpec {
		for _, cs := range tables {
			if cs.Name == name {
				return cs
			}
		}
		return nil
	}
	cmd.funcs["referenceFields"] = ReferenceFields
	/*
			cls
			var refCls *types.ClassSpec
			for _, cs := range tables {
				if cs.Name == name {
					refCls = cs
					break
				}
			}
			if refCls == nil {
				panic(errors.New("referenceFields of '" + f.Name + "' isn't string array in the " + cls.Name))
			}

			var fields = make([]types.FieldSpec, 0, len(names))
			for _, name := range names {
				found := false
				for _, field := range cls.Fields {
					if field.Name == fieldName {
						found = true
						fields = append(fields, field)
						break
					}
				}
				if !found {
					panic(errors.New("field '" + name + "' isn't exists in the " + cls.Name))
				}
			}
			return fields
		}
	*/

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
		if os.IsExist(err) {
			fmt.Println("[WARN] [EXISTS] skip", fname)
			return nil
		}
		return err
	}
	defer out.Close()

	for _, name := range names {
		tpl, err := cmd.newTemplate(name, funcs)
		if nil != err {
			return errors.New("load '" + name + "' template," + err.Error())
		}

		if err := tpl.Execute(out, params); err != nil {
			out.Close()
			os.Remove(fname)
			return errors.New("execute '" + name + "' template," + err.Error())
		}
	}
	return nil
}

type ReferenceField struct {
	Name  string
	Label string
}

func ReferenceFields(f types.FieldSpec) []ReferenceField {
	if f.Annotations == nil {
		return nil
	}

	a := f.Annotations["referenceFields"]
	if a == nil {
		return nil
	}
	var names []ReferenceField
	switch values := a.(type) {
	case []string:
		for _, s := range values {
			names = append(names, ReferenceField{Name: s})
		}
	case []interface{}:
		names = make([]ReferenceField, 0, len(values))
		for _, v := range values {
			switch value := v.(type) {
			case string:
				names = append(names, ReferenceField{Name: value})
			case map[string]interface{}:
				field := ReferenceField{Name: fmt.Sprint(value["name"])}
				if label := value["label"]; label != nil {
					field.Label = fmt.Sprint(label)
				}
				names = append(names, field)
			case map[interface{}]interface{}:
				field := ReferenceField{Name: fmt.Sprint(value["name"])}
				for k, vv := range value {
					switch fmt.Sprint(k) {
					case "name":
						field.Name = fmt.Sprint(vv)
					case "label":
						field.Label = fmt.Sprint(vv)
					}
				}
				names = append(names, field)
			default:
				fmt.Printf("%T %v\r\n", v, v)
				panic(errors.New("referenceFields of '" + f.Name + "' isn't string array or object array"))
			}
		}
	default:
		fmt.Printf("%T %v\r\n", a, a)
		panic(errors.New("referenceFields of '" + f.Name + "' isn't string array or object array"))
	}

	return names
}

func toFormatFunc(f types.FieldSpec) string {
	if f.Annotations != nil {
		if format := f.Annotations["columnFormat"]; format != nil {
			return fmt.Sprint(format)
		}

		if format := f.Annotations["enumerationSource"]; format != nil {
			return fmt.Sprint(format) + "_format"
		}
	}

	if f.Restrictions != nil && len(f.Restrictions.Enumerations) > 0 {
		return f.Name + "_format"
	}
	switch f.Format {
	case "net.IP", "ip", "mac", "email":
		return ""
	case "":
		return ""
	default:
		return f.Format
	}
}
