package types

type PropertySpec interface {
	PName() string
	TypeSpec() TypeDefinition
	CollectionType() CollectionType
	Required() bool
	ReadOnly() bool
	Uniquely() bool
	Validators() []Validator
	Default() interface{}
}

type KeySpec []PropertySpec

type ClassSpec interface {
	CName() string // CamelName
	UName() string // UnderscoreName

	Abstractly() bool
	IsInheritanced() bool
	ParentSpec() ClassSpec
	RootSpec() ClassSpec
	IsAssignableTo(super ClassSpec) bool
	IsInheritancedFrom(super ClassSpec) bool

	GetKeys() []KeySpec
	GetOwnProperties() []PropertySpec
	GetProperties() []PropertySpec
	GetProperty(nm string) PropertySpec
	GetOwnProperty(nm string) PropertySpec
	String() string
}

type SpecSystem struct {
	Underscore2Definitions map[string]ClassSpec
	Camel2Definitions      map[string]ClassSpec
}

func (self *SpecSystem) FindByUnderscoreName(nm string) ClassSpec {
	return self.Underscore2Definitions[nm]
}

func (self *SpecSystem) Find(nm string) ClassSpec {
	return self.Camel2Definitions[nm]
}

func (self *SpecSystem) Register(cls ClassSpec) {
	if _, ok := self.Camel2Definitions[cls.CName()]; ok {
		panic("class '" + cls.CName() + "' is already exists.")
	}
	if _, ok := self.Underscore2Definitions[cls.UName()]; ok {
		panic("class '" + cls.UName() + "' is already exists.")
	}
	self.Camel2Definitions[cls.CName()] = cls
	self.Underscore2Definitions[cls.UName()] = cls
}

func (self *SpecSystem) Unregister(cls ClassSpec) {
	delete(self.Camel2Definitions, cls.CName())
	delete(self.Underscore2Definitions, cls.UName())
}

func (self *SpecSystem) All() map[string]ClassSpec {
	return self.Camel2Definitions
}

func IsTable(spec ClassSpec) bool {
	_, ok := spec.(*TableDefinition)
	return ok
}

func IsDynamic(spec ClassSpec) bool {
	return spec.IsAssignableTo(DynamicClass)
}

var (
	DynamicClass = &ClassDefinition{
		BaseDefinition: BaseDefinition{
			Name:           "dynamic",
			UnderscoreName: "dynamic",
			IsAbstractly:   true,
			OwnFields:      map[string]*PropertyDefinition{},
			Fields:         map[string]*PropertyDefinition{},
		},
	}

	IntegerClass = &ClassDefinition{
		BaseDefinition: BaseDefinition{
			Name:           "Integer",
			UnderscoreName: "integer",
			IsAbstractly:   false,
			OwnFields:      map[string]*PropertyDefinition{},
			Fields:         map[string]*PropertyDefinition{},
		},
	}

	BigIntegerClass = &ClassDefinition{
		BaseDefinition: BaseDefinition{
			Name:           "BigInteger",
			UnderscoreName: "biginteger",
			IsAbstractly:   false,
			OwnFields:      map[string]*PropertyDefinition{},
			Fields:         map[string]*PropertyDefinition{},
		},
	}

	StringClass = &ClassDefinition{
		BaseDefinition: BaseDefinition{
			Name:           "String",
			UnderscoreName: "string",
			IsAbstractly:   false,
			OwnFields:      map[string]*PropertyDefinition{},
			Fields:         map[string]*PropertyDefinition{},
		},
	}

	DecimalClass = &ClassDefinition{
		BaseDefinition: BaseDefinition{
			Name:           "Decimal",
			UnderscoreName: "decimal",
			IsAbstractly:   false,
			OwnFields:      map[string]*PropertyDefinition{},
			Fields:         map[string]*PropertyDefinition{},
		},
	}

	DatetimeClass = &ClassDefinition{
		BaseDefinition: BaseDefinition{
			Name:           "Datetime",
			UnderscoreName: "datetime",
			IsAbstractly:   false,
			OwnFields:      map[string]*PropertyDefinition{},
			Fields:         map[string]*PropertyDefinition{},
		},
	}

	ObjectIdClass = &ClassDefinition{
		BaseDefinition: BaseDefinition{
			Name:           "ObjectId",
			UnderscoreName: "objectId",
			IsAbstractly:   false,
			OwnFields:      map[string]*PropertyDefinition{},
			Fields:         map[string]*PropertyDefinition{},
		},
	}

	IpAddressClass = &ClassDefinition{
		BaseDefinition: BaseDefinition{
			Name:           "IPAddress",
			UnderscoreName: "ipAddress",
			IsAbstractly:   false,
			OwnFields:      map[string]*PropertyDefinition{},
			Fields:         map[string]*PropertyDefinition{},
		},
	}

	PhysicalAddressClass = &ClassDefinition{
		BaseDefinition: BaseDefinition{
			Name:           "PhysicalAddress",
			UnderscoreName: "physicalAddress",
			IsAbstractly:   false,
			OwnFields:      map[string]*PropertyDefinition{},
			Fields:         map[string]*PropertyDefinition{},
		},
	}
)
