package types

import (
	"bytes"
)

type Annotation struct {
	Name  string
	Value interface{}
}

func mergeAnnotations(a, b []Annotation) []Annotation {
	res := make([]Annotation, 0, len(a)+len(b))
	return append(append(res, a...), b...)
}

func isEmbeddedAnnotation(ann Annotation) bool {
	return "embedded" == ann.Name
}

var EmbeddedAnnotation = Annotation{Name: "embedded"}

type CollectionType int

const (
	COLLECTION_UNKNOWN CollectionType = 0
	COLLECTION_ARRAY   CollectionType = 1
	COLLECTION_SET     CollectionType = 2
)

func (t CollectionType) IsArray() bool {
	return t == COLLECTION_ARRAY
}

func (t CollectionType) IsSet() bool {
	return t == COLLECTION_SET
}

func (t CollectionType) IsCollection() bool {
	return t == COLLECTION_SET || t == COLLECTION_ARRAY
}

type PropertyDefinition struct {
	Name         string
	Type         TypeDefinition
	Collection   CollectionType
	IsRequired   bool
	IsReadOnly   bool
	IsUniquely   bool
	Restrictions []Validator
	DefaultValue interface{}
	Annotations  []Annotation
}

func (self *PropertyDefinition) IsEmbedded() bool {
	for _, ann := range self.Annotations {
		if isEmbeddedAnnotation(ann) {
			return true
		}
	}
	return false
}

func (self *PropertyDefinition) IsSerial() bool {
	return "id" == self.Name
}

func (self *PropertyDefinition) IsPrimaryKey() bool {
	return "id" == self.Name
}

func (p *PropertyDefinition) PName() string {
	return p.Name
}
func (p *PropertyDefinition) TypeSpec() TypeDefinition {
	return p.Type
}
func (p *PropertyDefinition) CollectionType() CollectionType {
	return p.Collection
}
func (p *PropertyDefinition) Required() bool {
	return p.IsRequired
}
func (p *PropertyDefinition) ReadOnly() bool {
	return p.IsReadOnly
}
func (p *PropertyDefinition) Uniquely() bool {
	return p.IsUniquely
}
func (p *PropertyDefinition) Validators() []Validator {
	return p.Restrictions
}
func (p *PropertyDefinition) Default() interface{} {
	return p.DefaultValue
}

type BaseDefinition struct {
	Name           string
	UnderscoreName string
	CollectionName string
	IsAbstractly   bool
	Keys           [][]*PropertyDefinition
	OwnFields      map[string]*PropertyDefinition
	Fields         map[string]*PropertyDefinition
}

func (self *BaseDefinition) CName() string { // CamelName
	return self.Name
}
func (self *BaseDefinition) UName() string { // UnderscoreName
	return self.UnderscoreName
}
func (self *BaseDefinition) Abstractly() bool {
	return self.IsAbstractly
}

func (self *BaseDefinition) GetKeys() []KeySpec {
	ret := make([]KeySpec, 0, len(self.Keys))
	for _, fields := range self.Keys {
		key := make([]PropertySpec, 0, len(fields))
		for _, field := range fields {
			key = append(key, field)
		}
		ret = append(ret, KeySpec(key))
	}
	return ret
}

func (self *BaseDefinition) GetOwnProperties() []PropertySpec {
	ret := make([]PropertySpec, 0, len(self.OwnFields))
	for _, v := range self.OwnFields {
		ret = append(ret, v)
	}
	return ret
}

func (self *BaseDefinition) GetProperties() []PropertySpec {
	ret := make([]PropertySpec, 0, len(self.Fields))
	for _, v := range self.Fields {
		ret = append(ret, v)
	}
	return ret
}

func (self *BaseDefinition) GetProperty(nm string) PropertySpec {
	column := self.Fields[nm]
	if nil != column {
		return column
	}
	return nil
}

func (self *BaseDefinition) GetOwnProperty(nm string) PropertySpec {
	column := self.OwnFields[nm]
	if nil != column {
		return column
	}
	return nil
}

type ClassDefinition struct {
	BaseDefinition

	Super       *ClassDefinition
	OwnChildren []*ClassDefinition
}

func (self *ClassDefinition) IsInheritanced() bool {
	return (nil != self.Super) || (nil != self.OwnChildren && 0 != len(self.OwnChildren))
}
func (self *ClassDefinition) ParentSpec() ClassSpec {
	if nil == self.Super {
		return nil
	}
	return self.Super
}
func (self *ClassDefinition) RootSpec() ClassSpec {
	return self.Root()
}
func (self *ClassDefinition) Root() *ClassDefinition {
	s := self
	for nil != s.Super {
		s = s.Super
	}
	return s
}
func (self *ClassDefinition) IsAssignableTo(super ClassSpec) bool {
	parent, ok := super.(*ClassDefinition)
	if !ok {
		return false
	}
	return self == parent || self.IsSubclassOf(parent)
}
func (self *ClassDefinition) IsInheritancedFrom(super ClassSpec) bool {
	parent, ok := super.(*ClassDefinition)
	if !ok {
		return false
	}
	return self.IsSubclassOf(parent)
}
func (self *ClassDefinition) IsSubclassOf(cls *ClassDefinition) bool {
	for s := self; nil != s; s = s.Super {
		if s == cls {
			return true
		}
	}
	return false
}

func (self *ClassDefinition) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("class ")
	buffer.WriteString(self.Name)
	if nil != self.Super {
		buffer.WriteString(" < ")
		buffer.WriteString(self.Super.Name)
		buffer.WriteString(" { ")
	} else {
		buffer.WriteString(" { ")
	}
	if nil != self.OwnFields && 0 != len(self.OwnFields) {
		for _, pr := range self.OwnFields {
			buffer.WriteString(pr.Name)
			buffer.WriteString(",")
		}
		buffer.Truncate(buffer.Len() - 1)
	}
	buffer.WriteString(" }")
	return buffer.String()
}

type ClassDefinitions struct {
	underscore2Definitions map[string]*ClassDefinition
	clsDefinitions         map[string]*ClassDefinition
}

func (self *ClassDefinitions) FindByUnderscoreName(nm string) *ClassDefinition {
	return self.underscore2Definitions[nm]
}

func (self *ClassDefinitions) Find(nm string) *ClassDefinition {
	return self.clsDefinitions[nm]
}

func (self *ClassDefinitions) Register(cls *ClassDefinition) {
	self.clsDefinitions[cls.Name] = cls
	self.underscore2Definitions[cls.UnderscoreName] = cls
}

func (self *ClassDefinitions) Unregister(cls *ClassDefinition) {
	delete(self.clsDefinitions, cls.Name)
	delete(self.underscore2Definitions, cls.UnderscoreName)
}

func (self *ClassDefinitions) All() map[string]*ClassDefinition {
	return self.clsDefinitions
}

func MakeClassDefinitions(capacity int) *ClassDefinitions {
	return &ClassDefinitions{clsDefinitions: make(map[string]*ClassDefinition, capacity),
		underscore2Definitions: make(map[string]*ClassDefinition, capacity)}
}
