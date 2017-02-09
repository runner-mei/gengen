package types

import (
	"encoding/xml"
)

type XMLClassDefinitions struct {
	XMLName      xml.Name             `xml:"http://schemas.hengwei.com.cn/tpt/1/metricDefinitions classDefinitions"`
	LastModified string               `xml:"lastModified,attr"`
	Definitions  []XMLClassDefinition `xml:"class"`
	Mixins       []XMLMixinDefinition `xml:"mixin"`
}

type XMLMixinDefinition struct {
	XMLName    xml.Name                `xml:"mixin"`
	Name       string                  `xml:"name,attr"`
	Properties []XMLPropertyDefinition `xml:"property"`
}

type XMLClassDefinition struct {
	XMLName             xml.Name                 `xml:"class"`
	Name                string                   `xml:"name,attr"`
	Base                string                   `xml:"base,attr,omitempty"`
	Abstract            string                   `xml:"abstract,attr,omitempty"`
	Includes            []string                 `xml:"include,omitempty"`
	Labels              []XMLabel                `xml:"label,omitempty"`
	CombinedKeys        []XMLKey                 `xml:"combinedKey"`
	Properties          []XMLPropertyDefinition  `xml:"property"`
	BelongsTo           []XMLBelongsTo           `xml:"belongs_to"`
	HasMany             []XMLHasMany             `xml:"has_many"`
	HasOne              []XMLHasOne              `xml:"has_one"`
	HasAndBelongsToMany []XMLHasAndBelongsToMany `xml:"has_and_belongs_to_many"`
}

type XMLKey struct {
	Names []string `xml:"ref,omitempty"`
}

type XMLBelongsTo struct {
	Name   string `xml:"name,attr,omitempty"`
	Target string `xml:",chardata"`
}

type XMLHasMany struct {
	AttributeName string `xml:"attributeName,attr,omitempty"`
	ForeignKey    string `xml:"foreignKey,attr,omitempty"`
	Embedded      string `xml:"embedded,attr,omitempty"`
	Polymorphic   string `xml:"polymorphic,attr,omitempty"`
	Target        string `xml:",chardata"`
}

type XMLHasOne struct {
	AttributeName string `xml:"attributeName,attr,omitempty"`
	ForeignKey    string `xml:"foreignKey,attr,omitempty"`
	Embedded      string `xml:"embedded,attr,omitempty"`
	Target        string `xml:",chardata"`
}

type XMLHasAndBelongsToMany struct {
	ForeignKey string `xml:"foreignKey,attr,omitempty"`
	Through    string `xml:"through,attr,omitempty"`
	Target     string `xml:",chardata"`
}

type XMLPropertyDefinition struct {
	Name       string `xml:"name,attr"`
	Type       string `xml:"type,attr"`
	Embedded   string `xml:"embedded,attr,omitempty"`
	Collection string `xml:"collection,attr,omitempty"`

	Labels       *[]XMLabel   `xml:"label,omitempty"`
	Unit         string       `xml:"unit,omitempty"`
	Key          *XMLIsKey    `xml:",omitempty"`
	ReadOnly     *XMLReadOnly `xml:",omitempty"`
	Unique       *XMLUnique   `xml:",omitempty"`
	Required     *XMLRequired `xml:",omitempty"`
	DefaultValue string       `xml:"defaultValue,omitempty"`
	Enumerations *[]string    `xml:"enumeration>value,omitempty"`
	Pattern      string       `xml:"pattern,omitempty"`
	MinValue     string       `xml:"minValue,omitempty"`
	MaxValue     string       `xml:"maxValue,omitempty"`
	Length       string       `xml:"length,omitempty"`
	MinLength    string       `xml:"minLength,omitempty"`
	MaxLength    string       `xml:"maxLength,omitempty"`
}

// type XMLRestrictionsDefinition struct {
// 	XMLName xml.Name `xml:"restriction"`
// }

type XMLRequired struct {
	XMLName xml.Name `xml:"required"`
}

type XMLReadOnly struct {
	XMLName xml.Name `xml:"readonly"`
}

type XMLUnique struct {
	XMLName xml.Name `xml:"unique"`
}

type XMLIsKey struct {
	XMLName xml.Name `xml:"key"`
}

type XMLabel struct {
	XMLName xml.Name `xml:"label"`
	Lang    string   `xml:"lang,attr"`
	Value   string   `xml:",chardata"`
}
