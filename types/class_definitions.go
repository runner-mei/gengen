package types

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
)

func loadParentFields(classSpec ClassSpec, errs *[]string) {
	var definition, superDefinition *BaseDefinition
	var super ClassSpec
	switch t := classSpec.(type) {
	case *ClassDefinition:
		definition = &t.BaseDefinition
		if t.Super != nil {
			superDefinition = &t.Super.BaseDefinition
			super = t.Super
		}
	case *TableDefinition:
		definition = &t.BaseDefinition
		if t.Super != nil {
			superDefinition = &t.Super.BaseDefinition
			super = t.Super
		}
	default:
		panic(fmt.Errorf("class '%T' is unsupported", classSpec))
	}
	if nil != definition.Fields {
		return
	}

	definition.Fields = make(map[string]*PropertyDefinition, 2*len(definition.OwnFields))
	if nil != super {
		loadParentFields(super, errs)
		for k, v := range superDefinition.Fields {
			definition.Fields[k] = v
		}
	}

	for k, v := range definition.OwnFields {
		old, ok := definition.Fields[k]
		if ok {
			if v.Type != old.Type {
				*errs = append(*errs, "The property with '"+k+
					"' override failed, type is not same, own is '"+
					v.Type.Name()+"', super is '"+old.Type.Name()+"'")
			}

			// merge restrictions
			if nil != old.Restrictions {
				if nil == v.Restrictions {
					v.Restrictions = make([]Validator, 0)
				}

				for _, r := range old.Restrictions {
					v.Restrictions = append(v.Restrictions, r)
				}
			}

			// merge defaultValue
			if nil == v.DefaultValue {
				v.DefaultValue = old.DefaultValue
			}

			// merge isRequired
			if !v.IsRequired {
				v.IsRequired = old.IsRequired
			}

			v.Annotations = mergeAnnotations(v.Annotations, old.Annotations)
		}
		definition.Fields[k] = v
	}
}

func mergeErrors(errs []string, title string, msgs []string) []string {
	if nil == msgs || 0 == len(msgs) {
		return errs
	}
	if "" == title {
		for _, msg := range msgs {
			errs = append(errs, msg)
		}
		return errs
	}

	errs = append(errs, title)
	for _, msg := range msgs {
		errs = append(errs, "    "+msg)
	}
	return errs
}

func loadMixinFields(mixinDefinition *XMLMixinDefinition) (fields []*PropertyDefinition, errs []string) {
	fields = make([]*PropertyDefinition, 0, len(mixinDefinition.Properties))
	for _, pr := range mixinDefinition.Properties {
		if "type" == pr.Name {
			errs = append(errs, "load property '"+pr.Name+"' of mixin '"+
				mixinDefinition.Name+"' failed, it is reserved")
			continue
		}

		cpr, msgs := loadOwnField(&pr)
		if nil != cpr {
			fields = append(fields, cpr)
		}

		if nil != msgs && 0 != len(msgs) {
			errs = mergeErrors(errs, "load property '"+pr.Name+"' of mixin '"+
				mixinDefinition.Name+"' failed", msgs)
		}
	}
	return fields, errs
}

func loadOwnFields(xmlCls *XMLClassDefinition,
	cls *BaseDefinition) (errs []string) {
	cls.OwnFields = make(map[string]*PropertyDefinition)
	for _, pr := range xmlCls.Properties {
		// if "type" == pr.Name {
		// 	errs = append(errs, "load property '"+pr.Name+"' of class '"+
		// 		xmlCls.Name+"' failed, it is reserved")
		// 	continue
		// }

		var cpr *PropertyDefinition = nil
		cpr, msgs := loadOwnField(&pr)
		if nil != cpr {
			cls.OwnFields[cpr.Name] = cpr
		}

		errs = mergeErrors(errs, "load property '"+pr.Name+"' of class '"+
			xmlCls.Name+"' failed", msgs)
	}
	return errs
}

func loadOwnField(pr *XMLPropertyDefinition) (cpr *PropertyDefinition, errs []string) {
	if "" == pr.Type {
		errs = append(errs, "'type' in the '"+pr.Name+"' is required")
		return
	}
	if "_index" == pr.Name {
		errs = append(errs, "'_index' is reserved key.")
		return
	}
	cpr = &PropertyDefinition{Name: pr.Name,
		IsRequired:   false,
		Type:         GetTypeDefinition(pr.Type),
		Restrictions: make([]Validator, 0, 4)}

	if nil == cpr.Type {
		if "" == pr.Type {
			errs = append(errs, "type is empty.")
		} else {
			errs = append(errs, "'"+pr.Type+
				"' is unsupported type")
		}
		return nil, errs
	}

	switch pr.Collection {
	case "array":
		cpr.Collection = COLLECTION_ARRAY
	case "set":
		cpr.Collection = COLLECTION_SET
	default:
		cpr.Collection = COLLECTION_UNKNOWN
	}

	if "created_at" == pr.Name || "updated_at" == pr.Name {
		if "datetime" != cpr.Type.Name() {
			errs = append(errs, "it is reserved and must is a datetime")
		}

		if cpr.Collection.IsCollection() {
			errs = append(errs, "it is reserved and must not is a collection")
		}
	}

	if nil != pr.Required {
		cpr.IsRequired = true
	}

	if nil != pr.ReadOnly {
		cpr.IsReadOnly = true
	}

	if nil != pr.Unique {
		cpr.IsUniquely = true
	}

	if "" != pr.DefaultValue {
		if COLLECTION_UNKNOWN != cpr.Collection {
			errs = append(errs, "collection has not defaultValue ")
		} else {
			var err error
			cpr.DefaultValue, err = cpr.Type.ToInternal(pr.DefaultValue)
			if nil != err {
				errs = append(errs, "parse defaultValue '"+
					pr.DefaultValue+"' failed, "+err.Error())
			}
		}
	}

	if nil != pr.Enumerations && 0 != len(*pr.Enumerations) {
		validator, err := cpr.Type.CreateEnumerationValidator(*pr.Enumerations)
		if nil != err {
			errs = append(errs, "parse Enumerations '"+
				strings.Join(*pr.Enumerations, ",")+"' failed, "+err.Error())
		} else {
			cpr.Restrictions = append(cpr.Restrictions, validator)
		}
	}
	if "" != pr.Pattern {
		validator, err := cpr.Type.CreatePatternValidator(pr.Pattern)
		if nil != err {
			errs = append(errs, "parse Pattern '"+
				pr.Pattern+"' failed, "+err.Error())
		} else {
			cpr.Restrictions = append(cpr.Restrictions, validator)
		}
	}
	if "" != pr.MinValue || "" != pr.MaxValue {
		validator, err := cpr.Type.CreateRangeValidator(pr.MinValue,
			pr.MaxValue)
		if nil != err {
			errs = append(errs, "parse Range of Value '"+
				pr.MinValue+","+pr.MaxValue+
				"' failed, "+err.Error())
		} else {
			cpr.Restrictions = append(cpr.Restrictions, validator)
		}
	}
	if "" != pr.Length {
		validator, err := cpr.Type.CreateLengthValidator(pr.Length,
			pr.Length)
		if nil != err {
			errs = append(errs, "parse Length '"+
				pr.Length+"' failed, "+err.Error())
		} else {
			cpr.Restrictions = append(cpr.Restrictions, validator)
		}
	}
	if "" != pr.MinLength || "" != pr.MaxLength {
		validator, err := cpr.Type.CreateLengthValidator(pr.MinLength,
			pr.MaxLength)
		if nil != err {
			errs = append(errs, "parse Range of Length '"+
				pr.MinLength+","+pr.MaxLength+
				"' failed, "+err.Error())
		} else {
			cpr.Restrictions = append(cpr.Restrictions, validator)
		}
	}

	switch cpr.Name {
	case "type":
		if "string" != cpr.Type.Name() {
			errs = append(errs, "load column 'type' failed, it is reserved and must is a string")
		}

		if cpr.Collection.IsCollection() {
			errs = append(errs, "load column 'type' failed, it is reserved and must not is a collection")
		}
	case "record_version":
		if "integer" != cpr.Type.Name() {
			errs = append(errs, "load column 'record_version' failed, it is reserved and must is a integer")
		}

		if cpr.Collection.IsCollection() {
			errs = append(errs, "load column 'record_version' failed, it is reserved and must not is a collection")
		}

		errs = append(errs, "load column 'record_version' failed, it is reserved")
	}

	if cpr.IsEmbedded() {
		cpr.Annotations = append(cpr.Annotations, EmbeddedAnnotation)
	}

	if nil == errs || 0 == len(errs) {
		return cpr, nil
	}
	return nil, errs
}

func LoadClassDefinitionsFromFile(fileName string) (*ClassDefinitions, map[string][]*PropertyDefinition, error) {
	f, err := ioutil.ReadFile(fileName)
	if nil != err {
		return nil, nil, fmt.Errorf("read file '%s' failed, %s", fileName, err.Error())
	}

	var definitionList XMLClassDefinitions
	err = xml.Unmarshal(f, &definitionList)
	if nil != err {
		return nil, nil, fmt.Errorf("unmarshal xml '%s' failed, %s", fileName, err.Error())
	}
	return LoadClassDefinitions(fileName, definitionList.Definitions, definitionList.Mixins)
}

func LoadClassDefinitionsFromDir(dir string, pattern string) (*ClassDefinitions, map[string][]*PropertyDefinition, error) {
	if "" == pattern {
		pattern = "*.xml"
	}

	files, err := filepath.Glob(filepath.Join(dir, pattern))
	if nil != err {
		return nil, nil, fmt.Errorf("list dir '%s' failed, %s", dir, err.Error())
	}
	if nil == files || 0 == len(files) {
		return nil, nil, nil
	}

	definitions := make([]XMLClassDefinition, 0, 100)
	mixins := make([]XMLMixinDefinition, 0, 100)
	for _, fileName := range files {
		f, err := ioutil.ReadFile(fileName)
		if nil != err {
			return nil, nil, fmt.Errorf("read file '%s' failed, %s", fileName, err.Error())
		}

		var definitionList XMLClassDefinitions
		err = xml.Unmarshal(f, &definitionList)
		if nil != err {
			return nil, nil, fmt.Errorf("unmarshal xml '%s' failed, %s", fileName, err.Error())
		}

		definitions = append(definitions, definitionList.Definitions...)
		mixins = append(mixins, definitionList.Mixins...)
	}

	return LoadClassDefinitions(dir, definitions, mixins)
}

func loadMixinFieldDefinitions(namespace string, definitionList []XMLMixinDefinition, errs []string) (map[string][]*PropertyDefinition, []string) {
	if nil == definitionList || 0 == len(definitionList) {
		return nil, errs //fmt.Errorf("unmarshal xml '%s' error, mixin definition is empty", fileName)
	}

	mixinDefinitions := map[string][]*PropertyDefinition{}

	// load mixin definitions and own properties
	for _, xmlDefinition := range definitionList {
		name := xmlDefinition.Name
		if "" != namespace {
			name = namespace + ":" + xmlDefinition.Name
		}

		_, ok := mixinDefinitions[name]
		if ok {
			errs = append(errs, "mixin '"+name+
				"' is duplicated.")
			continue
		}

		fields, msgs := loadMixinFields(&xmlDefinition)
		if nil != fields {
			mixinDefinitions[name] = fields
		}
		if nil != msgs && 0 != len(msgs) {
			errs = mergeErrors(errs, "load mixin '"+name+"' failed", msgs)
		}
	}

	return mixinDefinitions, errs
}

func LoadClassDefinitions(fileName string, definitionList []XMLClassDefinition, mixin_definitions []XMLMixinDefinition) (*ClassDefinitions, map[string][]*PropertyDefinition, error) {
	self := &ClassDefinitions{clsDefinitions: make(map[string]*ClassDefinition, 100),
		underscore2Definitions: make(map[string]*ClassDefinition, 100)}
	self.Register(DynamicClass)
	self.Register(IntegerClass)
	self.Register(BigIntegerClass)
	self.Register(StringClass)
	self.Register(DecimalClass)
	self.Register(DatetimeClass)
	self.Register(ObjectIdClass)
	self.Register(IpAddressClass)
	self.Register(PhysicalAddressClass)

	return LoadClassDefinitionsWithOld(fileName, "", self, nil, definitionList, mixin_definitions)
}

func LoadClassDefinitionsWithOld(fileName, namespace string, self *ClassDefinitions, old_mixins map[string][]*PropertyDefinition, definitionList []XMLClassDefinition,
	mixin_definitions []XMLMixinDefinition) (*ClassDefinitions, map[string][]*PropertyDefinition, error) {
	if nil == definitionList || 0 == len(definitionList) {
		return self, old_mixins, nil //fmt.Errorf("unmarshal xml '%s' error, class definition is empty", fileName)
	}
	var mixins = map[string][]*PropertyDefinition{}
	var errs []string

	for k, v := range old_mixins {
		mixins[k] = v
	}

	mList, eList := loadMixinFieldDefinitions(namespace, mixin_definitions, make([]string, 0, 10))
	if len(mList) == 0 {
		for k, v := range mList {
			mixins[k] = v
		}
	}

	if len(eList) == 0 {
		errs = append(errs, eList...)
	}

	// load class definitions and own properties
	for _, xmlDefinition := range definitionList {
		className := xmlDefinition.Name
		if "" != namespace {
			className = namespace + ":" + xmlDefinition.Name
		}
		_, ok := self.clsDefinitions[className]
		if ok {
			errs = append(errs, "class '"+className+
				"' is duplicated.")
			continue
		}
		cls, msgs := loadClassDefinitionFirstStep(namespace, className, mixins, &xmlDefinition)
		if 0 == len(msgs) {
			errs = mergeErrors(errs, "load class '"+className+"' failed", msgs)
		}
		self.Register(cls)
	}

	// load super class
	for _, xmlDefinition := range definitionList {
		className := xmlDefinition.Name
		if "" != namespace {
			className = namespace + ":" + xmlDefinition.Name
		}

		cls, ok := self.clsDefinitions[className]
		if !ok {
			panic("'" + className + "' isn't found")
		}

		loadSuperClassWithOld(namespace, self, &xmlDefinition, cls, &errs)
	}

	// load the properties of super class
	for _, cls := range self.clsDefinitions {
		loadParentFields(cls, &errs)
	}

	if 0 == len(errs) {
		return self, mixins, nil
	}
	errs = mergeErrors(nil, "load file '"+fileName+"' error:", errs)
	return self, mixins, errors.New(strings.Join(errs, "\r\n"))
}

func loadSuperClassWithOld(namespace string, self *ClassDefinitions, xmlDefinition *XMLClassDefinition, cls *ClassDefinition, errs *[]string) {
	if "" == xmlDefinition.Base {
		return
	}

	var super *ClassDefinition
	var ok bool
	if "" != namespace {
		super, ok = self.clsDefinitions[namespace+":"+xmlDefinition.Base]
		if !ok || nil == super {
			super, ok = self.clsDefinitions[xmlDefinition.Base]
		}
	} else {
		super, ok = self.clsDefinitions[xmlDefinition.Base]
	}
	if !ok || nil == super {
		*errs = append(*errs, "Base '"+xmlDefinition.Base+
			"' of class '"+cls.Name+"' is not found.")
		return
	}

	if 0 == len(cls.Keys) {
		cls.Keys = super.Keys
	}
	cls.Super = super
	if nil == super.OwnChildren {
		super.OwnChildren = make([]*ClassDefinition, 0, 3)
	}
	super.OwnChildren = append(super.OwnChildren, cls)
}

func loadClassDefinitionFirstStep(namespace, classNameWithNs string, mixins map[string][]*PropertyDefinition, xmlDefinition *XMLClassDefinition) (*ClassDefinition, []string) {
	cls := &ClassDefinition{BaseDefinition: BaseDefinition{Name: classNameWithNs,
		UnderscoreName: Underscore(classNameWithNs)}}

	msgs := loadOwnFields(xmlDefinition, &cls.BaseDefinition)
	switch xmlDefinition.Abstract {
	case "true", "":
		cls.IsAbstractly = true
	case "false":
		cls.IsAbstractly = false
	default:
		msgs = append(msgs, "'abstract' value is invalid, it must is 'true' or 'false', actual is '"+xmlDefinition.Abstract+"'")
	}
	if nil != xmlDefinition.Includes && 0 != len(xmlDefinition.Includes) {
		if nil == mixins || 0 == len(mixins) {
			for _, include_mixin := range xmlDefinition.Includes {
				msgs = append(msgs, "mixin '"+include_mixin+"' isn't found.")
			}
		} else {
			for _, include_mixin := range xmlDefinition.Includes {
				var mixin []*PropertyDefinition
				var ok bool
				if "" != namespace {
					mixin, ok = mixins[namespace+":"+include_mixin]
					if !ok {
						mixin, ok = mixins[include_mixin]
					}
				} else {
					mixin, ok = mixins[include_mixin]
				}

				if !ok {
					msgs = append(msgs, "mixin '"+include_mixin+"' isn't found.")
					continue
				}
				if nil != mixin {
					for _, pr := range mixin {
						if _, found := cls.OwnFields[pr.Name]; found {
							msgs = append(msgs, "property '"+pr.Name+"' is duplicated.")
							continue
						}
						cls.OwnFields[pr.Name] = pr
					}
				}
			}
		}
	}
	for _, pr := range xmlDefinition.Properties {
		if nil != pr.Key {
			column, ok := cls.OwnFields[pr.Name]
			if !ok {
				panic("property '" + pr.Name + "' of '" + classNameWithNs + "' is not found.")
			}

			cls.Keys = append(cls.Keys, []*PropertyDefinition{column})
		}
	}

	for _, combinedKey := range xmlDefinition.CombinedKeys {
		if nil == combinedKey.Names || 0 == len(combinedKey.Names) {
			log.Println("[WARN] '" + classNameWithNs + "' has empty key.")
			continue
		}

		columns := make([]*PropertyDefinition, 0, len(combinedKey.Names))
		for _, nm := range combinedKey.Names {
			column, ok := cls.OwnFields[nm]
			if !ok {
				panic("property '" + nm + "' of '" + classNameWithNs + "' is not found.")
			}
			columns = append(columns, column)
		}
		cls.Keys = append(cls.Keys, columns)
	}
	return cls, msgs
}

func LoadClassDefinition(namespace string, self *ClassDefinitions,
	mixins map[string][]*PropertyDefinition, xmlDefinition *XMLClassDefinition) (*ClassDefinition, error) {
	className := xmlDefinition.Name
	if "" != namespace {
		className = namespace + ":" + xmlDefinition.Name
	}
	_, ok := self.clsDefinitions[className]
	if ok {
		return nil, errors.New("class '" + className + "' is duplicated.")
	}
	cls, msgs := loadClassDefinitionFirstStep(namespace, className, mixins, xmlDefinition)

	loadSuperClassWithOld(namespace, self, xmlDefinition, cls, &msgs)

	loadParentFields(cls, &msgs)
	if 0 != len(msgs) {
		return nil, errors.New("load class '" + className + "' failed: \r\n    " + strings.Join(msgs, "\r\n    "))
	}
	return cls, nil
}
