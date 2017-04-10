package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/runner-mei/gengen/types"
)

// GenerateViewCommand - 生成视图
type GenerateViewCommand struct {
	baseCommand
	layouts    string
	customPath string
}

// Flags - 申明参数
func (cmd *GenerateViewCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.layouts, "layouts", "", "")
	fs.StringVar(&cmd.customPath, "customPath", "", "")
	return cmd.baseCommand.Flags(fs)
}

// Run - 生成代码
func (cmd *GenerateViewCommand) Run(args []string) error {
	return cmd.run(args, cmd.genrateView)
}

func (cmd *GenerateViewCommand) genrateView(cls *types.ClassSpec) error {
	ctlName := Pluralize(cls.Name)
	params := map[string]interface{}{"namespace": cmd.ns,
		"controllerName": ctlName,
		"modelName":      ctlName,
		"layouts":        cmd.layouts,
		"customPath":     cmd.customPath,
		"class":          cls}
	funcs := template.FuncMap{"localizeName": localizeName,
		"index_label": func(cls *types.ClassSpec) string {
			if cls.IndexLabel != "" {
				return cls.IndexLabel
			}
			return localizeName(cls)
		},
		"new_label": func(cls *types.ClassSpec) string {
			if cls.NewLabel != "" {
				return cls.NewLabel
			}
			return "新建" + localizeName(cls)
		},
		"edit_label": func(cls *types.ClassSpec) string {
			if cls.EditLabel != "" {
				return cls.EditLabel
			}
			return "编辑" + localizeName(cls)
		},
		"isClob": func(f types.FieldSpec) bool {
			if f.Restrictions != nil {
				if f.Restrictions.Length > 500 {
					return true
				}
				if f.Restrictions.MaxLength > 500 {
					return true
				}
			}
			return false
		},
		"isID": func(f types.FieldSpec) bool {
			if f.Name == "id" {
				return true
			}
			return false
		},
		"needDisplay": func(f types.FieldSpec) bool {
			for k, ann := range f.Annotations {
				if k == "noshow" {
					if v := strings.ToLower(fmt.Sprint(ann)); v == "true" || v == "yes" {
						return false
					}
				}
			}

			if f.Name == "id" {
				return false
			}
			if f.Type == "password" {
				return false
			}
			return true
		},
		"jsEnumeration": func(enumerationValues []types.EnumerationValue) string {
			bs, err := json.Marshal(enumerationValues)
			if err != nil {
				panic(err)
			}
			return string(bs)
		},
		"hasEnumerations": func(f types.FieldSpec) bool {
			if f.Restrictions == nil {
				return false
			}
			return len(f.Restrictions.Enumerations) > 0
		},
		"belongsToClassName": func(cls *types.ClassSpec, f types.FieldSpec) string {
			for _, belongsTo := range cls.BelongsTo {
				if belongsTo.Name == f.Name {
					return belongsTo.Target
				}
			}
			return ""
		},
		"isBelongsTo": func(cls *types.ClassSpec, f types.FieldSpec) bool {
			for _, belongsTo := range cls.BelongsTo {
				if belongsTo.Name == f.Name {
					return true
				}
			}
			return false
		}}

	err := cmd.executeTempate(cmd.override, []string{"views/index"}, funcs, params, filepath.Join(cmd.output, ctlName, "index.html"))
	if err != nil {
		return errors.New("gen views/index: " + err.Error())
	}

	if !HasFeature(cls, "editDisabled") || !HasFeature(cls, "newDisabled") {
		err = cmd.executeTempate(cmd.override, []string{"views/fields"}, funcs, params, filepath.Join(cmd.output, ctlName, "edit_fields.html"))
		if err != nil {
			os.Remove(filepath.Join(cmd.output, "index.html"))
			os.Remove(filepath.Join(cmd.output, "edit.html"))
			return errors.New("gen views/fields: " + err.Error())
		}
		if !HasFeature(cls, "editDisabled") {
			err = cmd.executeTempate(cmd.override, []string{"views/edit"}, funcs, params, filepath.Join(cmd.output, ctlName, "edit.html"))
			if err != nil {
				os.Remove(filepath.Join(cmd.output, "index.html"))
				return errors.New("gen views/edit: " + err.Error())
			}
		}
		if !HasFeature(cls, "newDisabled") {
			err = cmd.executeTempate(cmd.override, []string{"views/new"}, funcs, params, filepath.Join(cmd.output, ctlName, "new.html"))
			if err != nil {
				os.Remove(filepath.Join(cmd.output, "index.html"))
				os.Remove(filepath.Join(cmd.output, "edit.html"))
				os.Remove(filepath.Join(cmd.output, "edit_fields.html"))
				return errors.New("gen views/new: " + err.Error())
			}
		}
	}

	err = cmd.executeTempate(cmd.override, []string{"views/quick"}, funcs, params, filepath.Join(cmd.output, ctlName, "quick-bar.html"))
	if err != nil {
		os.Remove(filepath.Join(cmd.output, "index.html"))
		os.Remove(filepath.Join(cmd.output, "edit.html"))
		os.Remove(filepath.Join(cmd.output, "edit_fields.html"))
		os.Remove(filepath.Join(cmd.output, "new.html"))
		return errors.New("gen views/quick: " + err.Error())
	}
	return nil
}

func localizeName(t interface{}) string {
	switch f := t.(type) {
	case types.FieldSpec:
		if f.Label != "" {
			return f.Label
		}
		return f.Name
	case *types.FieldSpec:
		if f.Label != "" {
			return f.Label
		}
		return f.Name
	case *types.ClassSpec:
		if f.Label != "" {
			return f.Label
		}
		return f.Name
	default:
		panic(fmt.Errorf("arguments of localizeName is unknown(%T: %v)", t, t))
	}
}

func HasFeature(f interface{}, name string) bool {
	var annotations map[string]interface{}
	switch v := f.(type) {
	case types.FieldSpec:
		annotations = v.Annotations
	case *types.FieldSpec:
		annotations = v.Annotations
	case *types.ClassSpec:
		annotations = v.Annotations
	case types.ClassSpec:
		annotations = v.Annotations
	default:
		panic(fmt.Errorf("unknown type - %T - %v", f, f))
	}

	for k, ann := range annotations {
		if k == name {
			if v := strings.ToLower(fmt.Sprint(ann)); v == "true" || v == "yes" {
				return true
			}
		}
	}
	return false
}
