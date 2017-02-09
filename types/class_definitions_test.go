package types

import (
	a "cn/com/hengwei/commons/assert"
	"net"
	"testing"
)

func TestLoadXml(t *testing.T) {
	definitions, _, err := LoadClassDefinitionsFromFile("test/test1.xml")
	if nil != err {
		t.Skip("read file 'test/test1.xml' failed, %s", err.Error())
		return
	}

	if nil == definitions.clsDefinitions {
		t.Skip("unmarshal xml 'test/test1.xml' error, classDefinition is nil")
		return
	}
	if 4 != len(definitions.clsDefinitions) {
		t.Skip("unmarshal xml 'test/test1.xml' error, len of classDefinitions is not 2", len(definitions.clsDefinitions))
		return
	}

	employee := definitions.Find("Employee")
	boss := definitions.Find("Boss")
	person := definitions.Find("Person")
	company := definitions.Find("Company")

	a.Check(t, boss.Super, a.Equals, employee, a.Commentf("check super of Class employee"))
	a.Check(t, employee.Super, a.Equals, person, a.Commentf("check super of Class employee"))
	a.Check(t, person.Super, a.IsNil, a.Commentf("check super of Class person"))
	a.Check(t, company.Super, a.IsNil, a.Commentf("check super of Class company"))

	a.Assert(t, len(employee.Fields), a.Equals, 12, a.Commentf("check len of Fields of employee"))
	a.Assert(t, len(person.Fields), a.Equals, 10, a.Commentf("check len of Fields of person"))
	a.Assert(t, len(company.Fields), a.Equals, 1, a.Commentf("check len of Fields of company"))

	assertProperty := func(p1, p2 *PropertyDefinition, comment int) {
		a.Check(t, p1.Name, a.Equals, p2.Name, a.Commentf("check Name of Fields[%d]", comment))
		a.Check(t, p1.Type, a.Equals, p2.Type,
			a.Commentf("check Type of Fields[%d]", comment))
		a.Check(t, p1.DefaultValue, a.DeepEquals, p1.DefaultValue,
			a.Commentf("check DefaultValue of Fields[%d]", comment))
		if nil == p1.Restrictions || 0 == len(p1.Restrictions) {
			if nil != p2.Restrictions && 0 != len(p1.Restrictions) {
				t.Errorf("check len(Restrictions) of Fields[%d]", comment)
			}
		} else if nil == p2.Restrictions || 0 == len(p2.Restrictions) {
			t.Errorf("check len(Restrictions) of Fields[%d]", comment)
		} else {
			a.Check(t, len(p1.Restrictions), a.Equals, len(p1.Restrictions),
				a.Commentf("check len(Restrictions) of Fields[%d]", comment))
		}
	}

	// assertBelongsTo := func(p1 Assocation, p2 *BelongsTo, comment int) {
	// 	a1 := p1.(*BelongsTo)
	// 	a.Check(t, a1, a.NotNil, a.Commentf("check Name of belongs_to[%d]", comment))
	// 	a.Check(t, a1.Name, a.Equals, p2.Name, a.Commentf("check Name of belongs_to[%d]", comment))
	// 	a.Check(t, p1.Target(), a.Equals, p2.Target(), a.Commentf("check Target of belongs_to[%d]", comment))
	// }

	// assertHasMany := func(p1 Assocation, p2 *HasMany, comment int) {
	// 	a1 := p1.(*HasMany)
	// 	a.Check(t, a1, a.NotNil, a.Commentf("check Name of has_many[%d]", comment))
	// 	a.Check(t, a1.ForeignKey, a.Equals, p2.ForeignKey, a.Commentf("check ForeignKey of has_many[%d]", comment))
	// 	a.Check(t, p1.Target(), a.Equals, p2.TargetClass, a.Commentf("check Target of has_many[%d]", comment))
	// }
	// assertHasOne := func(p1 Assocation, p2 *HasOne, comment int) {
	// 	a1 := p1.(*HasOne)
	// 	a.Check(t, a1, a.NotNil, a.Commentf("check Name of has_one[%d]", comment))
	// 	a.Check(t, a1.AttributeName, a.Equals, p2.AttributeName, a.Commentf("check AttributeName of has_one[%d]", comment))
	// 	a.Check(t, p1.Target(), a.Equals, p2.TargetClass, a.Commentf("check Target of has_one[%d]", comment))
	// }

	assertProperty(person.Fields["ID1"], &PropertyDefinition{Name: "ID1",
		Type:         GetTypeDefinition("integer"),
		DefaultValue: "0"}, 0)
	assertProperty(person.Fields["Name"], &PropertyDefinition{Name: "Name",
		Type:         GetTypeDefinition("string"),
		DefaultValue: "mfk",
		Restrictions: []Validator{&StringLengthValidator{
			MinLength: 3, MaxLength: 13}}}, 1)
	assertProperty(person.Fields["Name2"], &PropertyDefinition{Name: "Name2",
		Type:         GetTypeDefinition("string"),
		DefaultValue: "mfk",
		Restrictions: []Validator{&StringLengthValidator{
			MinLength: 3, MaxLength: 3}}}, 2)
	assertProperty(person.Fields["Age"], &PropertyDefinition{Name: "Age",
		Type:         GetTypeDefinition("integer"),
		DefaultValue: "mfk",
		Restrictions: []Validator{&IntegerValidator{
			MinValue: 3, MaxValue: 313}}}, 3)

	dateValidator, _ := GetTypeDefinition("datetime").CreateRangeValidator("2009-12-11T10:23:23+06:00",
		"2009-12-13T12:23:23+06:00")

	assertProperty(person.Fields["Day"], &PropertyDefinition{Name: "Day",
		Type:         GetTypeDefinition("datetime"),
		DefaultValue: "2009-12-12T12:23:23Z08:00",
		Restrictions: []Validator{dateValidator}}, 4)
	assertProperty(person.Fields["Mony"], &PropertyDefinition{Name: "Mony",
		Type:         GetTypeDefinition("decimal"),
		DefaultValue: "1.3",
		Restrictions: []Validator{&DecimalValidator{
			MinValue: 1.0, MaxValue: 3.0}}}, 5)
	assertProperty(person.Fields["IP"], &PropertyDefinition{Name: "IP",
		Type:         GetTypeDefinition("ipAddress"),
		DefaultValue: net.ParseIP("12.12.12.12")}, 6)
	mac, _ := net.ParseMAC("12-12-12-12-12-12")
	assertProperty(person.Fields["MAC"], &PropertyDefinition{Name: "MAC",
		Type:         GetTypeDefinition("physicalAddress"),
		DefaultValue: mac}, 7)

	enumValidator, _ := GetTypeDefinition("string").CreateEnumerationValidator([]string{"male", "female"})

	assertProperty(person.Fields["Sex"], &PropertyDefinition{Name: "Sex",
		Type:         GetTypeDefinition("string"),
		DefaultValue: "male",
		Restrictions: []Validator{enumValidator}}, 8)
	assertProperty(person.Fields["Password"], &PropertyDefinition{Name: "Password",
		Type:         GetTypeDefinition("password"),
		DefaultValue: "mfk"}, 9)

	assertProperty(employee.Fields["Job"], &PropertyDefinition{Name: "Job",
		Type:         GetTypeDefinition("string"),
		DefaultValue: "developer"}, 0)
	assertProperty(employee.Fields["company_test_id"], &PropertyDefinition{Name: "company_test_id",
		Type: GetTypeDefinition("objectId")}, 0)

	a.Check(t, company.Name, a.Equals, "Company", a.Commentf("check Class name"))

	a.Assert(t, len(company.Fields), a.Equals, 1, a.Commentf("check len of Fields"))

	assertProperty(company.Fields["Name"], &PropertyDefinition{Name: "Name",
		Type:         GetTypeDefinition("string"),
		DefaultValue: "Sina"}, 0)

	// if 3 != len(xmlDefinitions.Definitions) {
	// 	t.Skip("", len(xmlDefinitions.Definitions))
	// 	return
	// }
	// assertBelongsTo(employee.Assocations[0], &BelongsTo{TargetClass: company, Name: employee.Fields["company_test_id"]}, 0)
	// assertHasMany(company.Assocations[0], &HasMany{TargetClass: employee, ForeignKey: "company_test_id"}, 0)
	// assertHasOne(company.Assocations[1], &HasOne{TargetClass: boss, AttributeName: "boss"}, 0)
}

func TestPropertyOverride(t *testing.T) {
	definitions, _, err := LoadClassDefinitionsFromFile("test/test_property_override.xml")
	if nil != err {
		t.Skip("read file 'test/test_property_override.xml' failed, %s", err.Error())
		return
	}

	if nil == definitions.clsDefinitions {
		t.Skip("unmarshal xml 'test/test_property_override.xml' error, classDefinition is nil")
		return
	}
	if 3 != len(definitions.clsDefinitions) {
		t.Skip("unmarshal xml 'test/test_property_override.xml' error, len of classDefinitions is not 2", len(definitions.clsDefinitions))
		return
	}

	employee := definitions.Find("Employee")
	boss := definitions.Find("Boss")
	manager := definitions.Find("Manager")

	a.Check(t, employee.Super, a.IsNil, a.Commentf("check super of Class employee"))
	a.Check(t, boss.Super, a.Equals, employee, a.Commentf("check super of Class boss"))
	a.Check(t, manager.Super, a.Equals, employee, a.Commentf("check super of Class manager"))

	employeep := employee.Fields["Job"]
	bossp := boss.Fields["Job"]
	managerp := manager.Fields["Job"]

	t.Log(employee.OwnChildren)
	a.Check(t, len(employee.OwnChildren), a.Equals, 2, a.Commentf("check the OwnChildren of employee"))
	a.Check(t, employee.OwnChildren[0], a.Equals, boss, a.Commentf("check the OwnChildren[0] of employee is boss"))
	a.Check(t, employee.OwnChildren[1], a.Equals, manager, a.Commentf("check the OwnChildren[0] of employee is manager"))

	a.Check(t, employeep, a.NotNil)
	a.Check(t, bossp, a.NotNil)
	a.Check(t, managerp, a.NotNil)

	a.Check(t, employeep.DefaultValue, a.DeepEquals, "developer", a.Commentf("check the defaultValue of employee"))
	a.Check(t, bossp.DefaultValue, a.DeepEquals, "boss", a.Commentf("check the defaultValue of boss"))
	a.Check(t, managerp.DefaultValue, a.DeepEquals, "manager", a.Commentf("check the defaultValue of manager"))

	if nil != employeep.Restrictions && 0 != len(employeep.Restrictions) {
		t.Skip("check the restrictions of employee")
	}
	a.Check(t, len(bossp.Restrictions), a.Equals, 1, a.Commentf("check the restrictions of boss"))
	a.Check(t, len(managerp.Restrictions), a.Equals, 1, a.Commentf("check the restrictions of manager"))

	a.Check(t, bossp.Restrictions[0], a.DeepEquals, &StringLengthValidator{MinLength: 3, MaxLength: 13}, a.Commentf("check the restrictions of boss"))
	a.Check(t, managerp.Restrictions[0], a.DeepEquals, &StringLengthValidator{MinLength: 4, MaxLength: 14}, a.Commentf("check the restrictions of manager"))
}

func TestHierarchicalTypeIsOk(t *testing.T) {
	t.Skip("not implemented")
}

func TestHierarchicalTypeIsInputError(t *testing.T) {
	t.Skip("not implemented")
}

func TestHierarchicalTypeWithParnetIsNilAndChildNotNil(t *testing.T) {
	t.Skip("not implemented")
}

func TestHierarchicalTypeWithParnetNotNilAndChildIsNil(t *testing.T) {
	t.Skip("not implemented")
}

func TestHierarchicalTypeWithChildMaxValueIsError(t *testing.T) {
	t.Skip("not implemented")
}

func TestHierarchicalTypeWithChildMinValueIsError(t *testing.T) {
	t.Skip("not implemented")
}

func TestHierarchicalTypeWithOwnChildrenIsOverlappingAtLeft(t *testing.T) {
	t.Skip("not implemented")
}

func TestHierarchicalTypeWithOwnChildrenIsOverlappingAtRight(t *testing.T) {
	t.Skip("not implemented")
}

func TestClassIsAleadyExists(t *testing.T) {
	t.Skip("not implemented")
}

func TestSuperClassNotFound(t *testing.T) {
	t.Skip("not implemented")
}

func TestTargetClassOfBelongsToNotFound(t *testing.T) {
	t.Skip("not implemented")
}

func TestNameOfBelongsToNotFound(t *testing.T) {
	t.Skip("not implemented")
}

func TestTargetClassOfHasManyNotFound(t *testing.T) {
	t.Skip("not implemented")
}

func TestTargetClassOfHasAndBelongsToManyNotFound(t *testing.T) {
	t.Skip("not implemented")
}

func TestPropertyTypeIsUnsupportedType(t *testing.T) {
	t.Skip("not implemented")
}

func TestLengthOfRestrictionIsError(t *testing.T) {
	t.Skip("not implemented")
}

func TestMaxLengthAndMinLengthOfRestrictionIsError(t *testing.T) {
	t.Skip("not implemented")
}

func TestMaxValueAndMinValueOfRestrictionIsError(t *testing.T) {
	t.Skip("not implemented")
}

func TestPatternOfRestrictionIsError(t *testing.T) {
	t.Skip("not implemented")
}

func TestEnumerationOfRestrictionIsError(t *testing.T) {
	t.Skip("not implemented")
}

func TestDefaultValueOfRestrictionIsError(t *testing.T) {
	t.Skip("not implemented")
}

func TestTypeOfPropertyIsMismatchInSuperAndChild(t *testing.T) {
	t.Skip("not implemented")
}
