package types

import (
	a "cn/com/hengwei/commons/assert"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"testing"
)

func TestOutXML(t *testing.T) {
	classes := &XMLClassDefinitions{
		LastModified: "123",
		Definitions:  make([]XMLClassDefinition, 0)}

	cl1_properties := []XMLPropertyDefinition{XMLPropertyDefinition{Name: "Id", Labels: &[]XMLabel{XMLabel{Lang: "zh", Value: "标识符"}}, Type: "integer", DefaultValue: "12"},
		XMLPropertyDefinition{Name: "Sex", Type: "string", Enumerations: &[]string{"male", "female"}},
		XMLPropertyDefinition{Name: "Name", Type: "string", Pattern: "a.*"},
		XMLPropertyDefinition{Name: "Age", Type: "integer", MinValue: "1", MaxValue: "130"},
		XMLPropertyDefinition{Name: "Address", Type: "string", MinLength: "10", MaxLength: "20"}}

	cl1 := &XMLClassDefinition{Name: "Person", Properties: cl1_properties}

	cl2 := &XMLClassDefinition{Name: "Employee", Base: "Person",
		BelongsTo:           []XMLBelongsTo{XMLBelongsTo{Name: "cc_id", Target: "CC"}, XMLBelongsTo{Name: "bb_id", Target: "BB"}},
		HasMany:             []XMLHasMany{XMLHasMany{Target: "DD"}, XMLHasMany{Target: "BB"}},
		HasOne:              []XMLHasOne{XMLHasOne{Target: "DD"}, XMLHasOne{Target: "BB"}},
		HasAndBelongsToMany: []XMLHasAndBelongsToMany{XMLHasAndBelongsToMany{Target: "DD"}, XMLHasAndBelongsToMany{Target: "BB"}},
		Properties: []XMLPropertyDefinition{
			XMLPropertyDefinition{Name: "Id2", Type: "string", DefaultValue: "12"},
			XMLPropertyDefinition{Name: "Sex2", Type: "string", Enumerations: &[]string{"male", "female"}},
			XMLPropertyDefinition{Name: "Name2", Type: "string", Pattern: "a.*"},
			XMLPropertyDefinition{Name: "Age2", Type: "string", MinValue: "1", MaxValue: "130"},
			XMLPropertyDefinition{Name: "Address2", Type: "string", MinLength: "10", MaxLength: "20"}}}

	classes.Definitions = append(classes.Definitions, *cl1, *cl2)

	output, err := xml.MarshalIndent(classes, "  ", "    ")
	if err != nil {
		t.Errorf("error: %v\n", err)
	}
	fmt.Println(xml.Header)
	fmt.Println(string(output))

	var xmlDefinitions XMLClassDefinitions
	err = xml.Unmarshal(output, &xmlDefinitions)
	if nil != err {
		t.Errorf("unmarshal xml failed, %s", err.Error())
		return
	}
	if nil == xmlDefinitions.Definitions {
		t.Errorf("unmarshal xml error, classDefinition is nil")
		return
	}
	if 2 != len(xmlDefinitions.Definitions) {
		t.Errorf("unmarshal xml error, len of classDefinitions is not 2", len(xmlDefinitions.Definitions))
		return
	}

	person := xmlDefinitions.Definitions[0]
	employee := xmlDefinitions.Definitions[1]

	a.Check(t, person.Name, a.Equals, "Person", a.Commentf("check Class name"))
	a.Check(t, person.Base, a.Equals, "", a.Commentf("check Base name"))
	a.Assert(t, len(person.Properties), a.Equals, 5, a.Commentf("check len of Properties"))
	a.Check(t, person.Properties[0].Name, a.Equals, "Id", a.Commentf("check name of Properties[0]"))

	a.Check(t, employee.Name, a.Equals, "Employee", a.Commentf("check Class name"))
	a.Check(t, employee.Base, a.Equals, "Person", a.Commentf("check Base name"))
}

// func TestXml(t *testing.T) {
// 	type Email struct {
// 		Where string `xml:"where,attr"`
// 		Addr  string
// 	}
// 	type Address struct {
// 		City, State string
// 	}
// 	type Result struct {
// 		XMLName xml.Name `xml:"Person"`
// 		Name    string   `xml:"FullName"`
// 		Phone   string
// 		Email   []Email  `xml:"email"`
// 		Groups  []string `xml:"Group>Value"`
// 		Address
// 	}
// 	v := Result{Name: "none", Phone: "none"}

// 	data := `
//     <Person>
//         <FullName>Grace R. Emlin</FullName>
//         <Company>Example Inc.</Company>
//         <email where="home">
//             <Addr>gre@example.com</Addr>
//         </email>
//         <email where='work'>
//             <Addr>gre@work.com</Addr>
//         </email>
//         <Group>
//             <Value>Friends</Value>
//             <Value>Squash</Value>
//         </Group>
//         <City>Hanga Roa</City>
//         <State>Easter Island</State>
//     </Person>
// `
// 	err := xml.Unmarshal([]byte(data), &v)
// 	if err != nil {
// 		fmt.Printf("error: %v", err)
// 		return
// 	}
// 	fmt.Printf("XMLName: %#v\n", v.XMLName)
// 	fmt.Printf("Name: %q\n", v.Name)
// 	fmt.Printf("Phone: %q\n", v.Phone)
// 	fmt.Printf("Email: %v\n", v.Email)
// 	fmt.Printf("Groups: %v\n", v.Groups)
// 	fmt.Printf("Address: %v\n", v.Address)
// }

func checkArray(s1, s2 *[]string) bool {
	if nil == s1 || nil == *s1 || 0 == len(*s1) {
		if nil == s2 || nil == *s2 || 0 == len(*s2) {
			return true
		}
		return false
	}

	if len(*s1) != len(*s2) {
		return false
	}

	for i, s := range *s1 {
		if s != (*s2)[i] {
			return false
		}
	}
	return true
}

// func TestXML2(t *testing.T) {
// 	bytes, err := ioutil.ReadFile("test/test_property_override.xml")
// 	if nil != err {
// 		t.Errorf("read file 'test/test1.xml' failed, %s", err.Error())
// 		return
// 	}
// 	var xmlDefinitions XMLClassDefinitions
// 	err = xml.Unmarshal(bytes, &xmlDefinitions)
// 	if nil != err {
// 		t.Errorf("unmarshal xml 'test/test1.xml' failed, %s", err.Error())
// 		return
// 	}
// 	if nil == xmlDefinitions.Definitions {
// 		t.Errorf("unmarshal xml 'test/test1.xml' error, classDefinition is nil")
// 		return
// 	}
// 	if 3 != len(xmlDefinitions.Definitions) {
// 		t.Errorf("unmarshal xml 'test/test1.xml' error, len of classDefinitions is not 2", len(xmlDefinitions.Definitions))
// 		return
// 	}

// 	output, err := xml.MarshalIndent(xmlDefinitions, "  ", "    ")
// 	if err != nil {
// 		t.Errorf("error: %v\n", err)
// 	}
// 	os.Stdout.Write([]byte(xml.Header))
// 	os.Stdout.Write(output)
// }

func TestXML1(t *testing.T) {
	bytes, err := ioutil.ReadFile("test/test1.xml")
	if nil != err {
		t.Errorf("read file 'test/test1.xml' failed, %s", err.Error())
		return
	}
	var xmlDefinitions XMLClassDefinitions
	err = xml.Unmarshal(bytes, &xmlDefinitions)
	if nil != err {
		t.Errorf("unmarshal xml 'test/test1.xml' failed, %s", err.Error())
		return
	}
	if nil == xmlDefinitions.Definitions {
		t.Errorf("unmarshal xml 'test/test1.xml' error, classDefinition is nil")
		return
	}
	if 4 != len(xmlDefinitions.Definitions) {
		t.Errorf("unmarshal xml 'test/test1.xml' error, len of classDefinitions is not 2", len(xmlDefinitions.Definitions))
		return
	}
	if nil == xmlDefinitions.Mixins {
		t.Errorf("unmarshal xml 'test/test1.xml' error, Mixins is nil")
		return
	}
	if 1 != len(xmlDefinitions.Mixins) {
		t.Errorf("unmarshal xml 'test/test1.xml' error, len of Mixins is not 1", len(xmlDefinitions.Mixins))
		return
	}

	person_mixin := xmlDefinitions.Mixins[0]
	if nil == person_mixin.Properties || 1 != len(person_mixin.Properties) {
		t.Errorf("unmarshal xml 'test/test1.xml' error, len of Properties is not 1", len(person_mixin.Properties))
		return
	}
	employee := xmlDefinitions.Definitions[0]
	boss := xmlDefinitions.Definitions[1]
	person := xmlDefinitions.Definitions[2]
	company := xmlDefinitions.Definitions[3]

	a.Check(t, person_mixin.Name, a.Equals, "PersonMixin", a.Commentf("check Class name of person"))
	a.Check(t, person.Name, a.Equals, "Person", a.Commentf("check Class name of person"))
	a.Check(t, person.Base, a.Equals, "", a.Commentf("check Base name of person"))
	a.Assert(t, len(person.Properties), a.Equals, 10, a.Commentf("check len of Properties of person"))
	a.Check(t, person.Properties[0].Name, a.Equals, "ID1", a.Commentf("check name of Properties[0] of person"))

	assertProperty := func(p1, p2 *XMLPropertyDefinition, comment int) {
		a.Check(t, p1.Name, a.Equals, p2.Name, a.Commentf("check Name of properties[%d]", comment))
		a.Check(t, p1.Type, a.Equals, p2.Type, a.Commentf("check Restrictions.Type of properties[%d] - %v", comment, p2.Name))
		a.Check(t, p1.DefaultValue, a.DeepEquals, p2.DefaultValue,
			a.Commentf("check Restrictions.DefaultValue of properties[%d]", comment))

		a.Check(t, p1.Required, a.DeepEquals, p2.Required,
			a.Commentf("check Restrictions.Required of properties[%d]", comment))

		if !checkArray(p1.Enumerations, p2.Enumerations) {
			t.Errorf("check Restrictions.Enumerations properties[%d] failed, value is %v", comment, p1.Enumerations)
		}

		a.Check(t, p1.Collection, a.Equals, p2.Collection,
			a.Commentf("check Restrictions.Collection properties[%d]", comment))
		a.Check(t, p1.Length, a.Equals, p2.Length,
			a.Commentf("check Restrictions.Length properties[%d]", comment))
		a.Check(t, p1.MaxLength, a.Equals, p2.MaxLength,
			a.Commentf("check Restrictions.MaxLength properties[%d]", comment))
		a.Check(t, p1.MinLength, a.Equals, p2.MinLength,
			a.Commentf("check Restrictions.MinLength properties[%d]", comment))
		a.Check(t, p1.MaxValue, a.Equals, p2.MaxValue,
			a.Commentf("check Restrictions.MaxValue properties[%d]", comment))
		a.Check(t, p1.MinValue, a.Equals, p2.MinValue,
			a.Commentf("check Restrictions.MinValue properties[%d]", comment))
		a.Check(t, p1.Pattern, a.Equals, p2.Pattern,
			a.Commentf("check Restrictions.Pattern properties[%d]", comment))
	}

	assertBelongsTo := func(p1, p2 *XMLBelongsTo, comment int) {
		a.Check(t, p1.Name, a.Equals, p2.Name, a.Commentf("check Name of belongs_to[%d]", comment))
		a.Check(t, p1.Target, a.Equals, p2.Target, a.Commentf("check Target of belongs_to[%d]", comment))
	}

	assertHasMany := func(p1, p2 *XMLHasMany, comment int) {
		a.Check(t, p1.ForeignKey, a.Equals, p2.ForeignKey, a.Commentf("check ForeignKey of has_many[%d]", comment))
		a.Check(t, p1.Target, a.Equals, p2.Target, a.Commentf("check Target of has_many[%d]", comment))
	}

	assertHasOne := func(p1, p2 *XMLHasOne, comment int) {
		a.Check(t, p1.AttributeName, a.Equals, p2.AttributeName, a.Commentf("check AttributeName of has_one[%d]", comment))
		a.Check(t, p1.Target, a.Equals, p2.Target, a.Commentf("check Target of has_one[%d]", comment))
	}

	assertProperty(&person_mixin.Properties[0], &XMLPropertyDefinition{Name: "ID3",
		Type: "integer", DefaultValue: "0"}, 0)
	assertProperty(&person.Properties[0], &XMLPropertyDefinition{Name: "ID1",
		Type: "integer", DefaultValue: "0"}, 0)
	assertProperty(&person.Properties[1], &XMLPropertyDefinition{Name: "Name",
		Type:         "string",
		DefaultValue: "mfk",
		MinLength:    "3", MaxLength: "13"}, 1)
	assertProperty(&person.Properties[2], &XMLPropertyDefinition{Name: "Name2",
		Type:         "string",
		DefaultValue: "mfk",
		Length:       "3"}, 2)
	assertProperty(&person.Properties[3], &XMLPropertyDefinition{Name: "Age",
		Type:         "integer",
		DefaultValue: "123",
		MinValue:     "3",
		MaxValue:     "313"}, 3)
	assertProperty(&person.Properties[4], &XMLPropertyDefinition{Name: "Day",
		Type:         "datetime",
		DefaultValue: "2009-12-12T12:23:23+08:00",
		MinValue:     "2009-12-11T10:23:23+08:00",
		MaxValue:     "2009-12-13T12:23:23+08:00"}, 4)
	assertProperty(&person.Properties[5], &XMLPropertyDefinition{Name: "Mony",
		Type:         "decimal",
		DefaultValue: "1.3",
		MinValue:     "1.0",
		MaxValue:     "3.0"}, 5)
	assertProperty(&person.Properties[6], &XMLPropertyDefinition{Name: "IP",
		Type:         "ipAddress",
		DefaultValue: "12.12.12.12"}, 6)
	assertProperty(&person.Properties[7], &XMLPropertyDefinition{Name: "MAC",
		Type:         "physicalAddress",
		DefaultValue: "12-12-12-12-12-12"}, 7)
	assertProperty(&person.Properties[8], &XMLPropertyDefinition{Name: "Sex",
		Type:         "string",
		DefaultValue: "male",
		Enumerations: &[]string{"male", "female"}}, 8)
	assertProperty(&person.Properties[9], &XMLPropertyDefinition{Name: "Password",
		Type:         "password",
		DefaultValue: "mfk"}, 9)

	a.Check(t, employee.Name, a.Equals, "Employee", a.Commentf("check Class name"))
	a.Check(t, employee.Base, a.Equals, "Person", a.Commentf("check Base name"))

	a.Assert(t, len(employee.Properties), a.Equals, 2, a.Commentf("check len of Properties"))

	assertProperty(&employee.Properties[0], &XMLPropertyDefinition{Name: "Job",
		Type:     "string",
		Required: &XMLRequired{XMLName: xml.Name{Space: "http://schemas.hengwei.com.cn/tpt/1/metricDefinitions", Local: "required"}}}, 0)
	assertProperty(&employee.Properties[1], &XMLPropertyDefinition{Name: "company_test_id",
		Type: "objectId"}, 1)

	a.Check(t, boss.Name, a.Equals, "Boss", a.Commentf("check Class name of boss"))
	a.Check(t, boss.Base, a.Equals, "Employee", a.Commentf("check Base name of boss"))

	a.Assert(t, len(boss.Properties), a.Equals, 1, a.Commentf("check len of Properties"))

	assertProperty(&boss.Properties[0], &XMLPropertyDefinition{Name: "Job",
		Type:         "string",
		DefaultValue: "boss", MinLength: "3", MaxLength: "13"}, 0)

	a.Check(t, company.Name, a.Equals, "Company", a.Commentf("check Class company.name"))

	a.Assert(t, len(company.Properties), a.Equals, 1, a.Commentf("check len of company.Properties"))

	assertProperty(&company.Properties[0], &XMLPropertyDefinition{Name: "Name",
		Type:         "string",
		DefaultValue: "Sina"}, 0)

	// if 3 != len(xmlDefinitions.Definitions) {
	// 	t.Errorf("", len(xmlDefinitions.Definitions))
	// 	return
	// }
	assertBelongsTo(&employee.BelongsTo[0], &XMLBelongsTo{Target: "Company", Name: "company_test_id"}, 0)
	assertHasMany(&company.HasMany[0], &XMLHasMany{Target: "Employee", ForeignKey: "company_test_id"}, 0)

	assertHasOne(&company.HasOne[0], &XMLHasOne{Target: "Boss", AttributeName: "boss"}, 0)
}
