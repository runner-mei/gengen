package types

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type ClassSpec struct {
	Name         string      `json:"name" yaml:"name"`
	Label        string      `json:"label,omitempty" ymal:"label,omitempty"`
	IndexLabel   string      `json:"index_label" yaml:"index_label"`
	NewLabel     string      `json:"new_label,omitempty" yaml:"new_label,omitempty"`
	EditLabel    string      `json:"edit_label,omitempty" yaml:"edit_label,omitempty"`
	Table        string      `json:"table,omitempty" yaml:"table,omitempty"`
	IsAbstractly bool        `json:"abstract,omitempty" yaml:"abstract,omitempty"`
	Keys         [][]string  `json:"keys,omitempty" yaml:"keys,omitempty"`
	Fields       []FieldSpec `json:"fields,omitempty" yaml:"fields,omitempty"`

	PrimaryKey          []string              `json:"primaryKey,omitempty" yaml:"primaryKey,omitempty"`
	HasMany             []HasMany             `json:"hasMany,omitempty" yaml:"hasMany,omitempty"`
	BelongsTo           []BelongsTo           `json:"belongsTo,omitempty" yaml:"belongsTo,omitempty"`
	HasAndBelongsToMany []HasAndBelongsToMany `json:"hasAndBelongsToMany,omitempty" yaml:"hasAndBelongsToMany,omitempty"`

	Annotations map[string]interface{} `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type HasMany struct {
	Target      string `json:"target" yaml:"target"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	ForeignKey  string `json:"foreignKey,omitempty" yaml:"foreignKey,omitempty"`
	Polymorphic string `json:"polymorphic,omitempty" yaml:"polymorphic,omitempty"`
}

type BelongsTo struct {
	Target string `json:"target" yaml:"target"`
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
}

func (belongsTo *BelongsTo) AttributeName(json bool) string {
	if belongsTo.Name != "" {
		if json {
			return Underscore(belongsTo.Name)
		}
		name := Goify(belongsTo.Name, true)
		if strings.HasSuffix(name, "Id") {
			return strings.TrimSuffix(name, "Id") + "ID"
		}
		return name
	}
	if json {
		return Underscore(belongsTo.Target) + "_id"
	}
	return Goify(Underscore(belongsTo.Target)+"_id", true)
}

type HasAndBelongsToMany struct {
	Target     string `json:"target" yaml:"target"`
	ForeignKey string `json:"foreignKey,omitempty" yaml:"foreignKey,omitempty"`
	Through    string `json:"through,omitempty" yaml:"through,omitempty"`
}

type FieldSpec struct {
	Name         string                 `json:"name" ymal:"name"`
	Label        string                 `json:"label,omitempty" ymal:"label,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	Type         string                 `json:"type" yaml:"type"`
	Format       string                 `json:"format" yaml:"format"`
	Collection   bool                   `json:"is_array,omitempty" yaml:"is_array,omitempty"`
	IsEmbedded   bool                   `json:"embedded,omitempty" yaml:"embedded,omitempty"`
	IsRequired   bool                   `json:"required,omitempty" yaml:"required,omitempty"`
	IsReadOnly   bool                   `json:"readonly,omitempty" yaml:"readonly,omitempty"`
	IsUniquely   bool                   `json:"unique,omitempty" yaml:"unique,omitempty"`
	DefaultValue string                 `json:"default,omitempty" yaml:"default,omitempty"`
	Unit         string                 `json:"unit,omitempty" yaml:"unit,omitempty"`
	Restrictions *RestrictionSpec       `json:"restrictions,omitempty" yaml:"restrictions,omitempty"`
	Annotations  map[string]interface{} `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type RestrictionSpec struct {
	Enumerations []EnumerationValue `json:"enumerations,omitempty" yaml:"enumerations,omitempty"`
	Pattern      string             `json:"pattern,omitempty" yaml:"pattern,omitempty"`
	MinValue     string             `json:"minValue,omitempty" yaml:"minValue,omitempty"`
	MaxValue     string             `json:"maxValue,omitempty" yaml:"maxValue,omitempty"`
	Length       int                `json:"length,omitempty" json:"length,omitempty"`
	MinLength    int                `json:"minLength,omitempty" json:"minLength,omitempty"`
	MaxLength    int                `json:"maxLength,omitempty" json:"maxLength,omitempty"`
}

type EnumerationValue struct {
	Label string `json:"label,omitempty" ymal:"label,omitempty"`
	Value string `json:"value,omitempty" ymal:"value,omitempty"`
}

func LoadYAMLFiles(filenames []string) ([]*ClassSpec, error) {
	var classList []*ClassSpec
	for _, filename := range filenames {
		cs, err := LoadYAMLFile(filename)
		if err != nil {
			return nil, err
		}
		classList = append(classList, cs)
	}
	return classList, nil
}

func LoadYAMLFile(filename string) (*ClassSpec, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cs = &ClassSpec{}
	if err = yaml.Unmarshal(bs, cs); err != nil {
		return nil, errors.New("load " + filename + " fail, " + err.Error())
	}
	return cs, nil
}

func LoadJSONFiles(filenames []string) ([]*ClassSpec, error) {
	var classList []*ClassSpec
	for _, filename := range filenames {
		cs, err := LoadJSONFile(filename)
		if err != nil {
			return nil, err
		}
		classList = append(classList, cs)
	}
	return classList, nil
}

func LoadJSONFile(filename string) (*ClassSpec, error) {
	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cs = &ClassSpec{}
	return cs, json.Unmarshal(bs, cs)
}

var ToGoTypes = map[string]string{"boolean": "bool",
	"integer":         "int",
	"decimal":         "float64",
	"string":          "string",
	"datetime":        "time.Time",
	"duration":        "time.Duration",
	"ipaddress":       "net.IP",
	"ipAddress":       "net.IP",
	"IPAddress":       "net.IP",
	"physicalAddress": "[]byte",
	"PhysicalAddress": "[]byte",
	"password":        "string",
	"objectId":        "int64",
	"objectID":        "int64",
	"biginteger":      "int64",
	"bigInteger":      "int64",
	"map":             "map[string]interface{}",
	"dynamic":         ""}
