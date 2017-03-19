package main

import (
	"cn/com/hengwei/commons/uuid"
	"crypto/md5"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/manveru/faker"
	"github.com/runner-mei/gengen/types"
)

// GenerateUnitTestCommand - 生成视图
type GenerateUnitTestCommand struct {
	baseCommand
	projectPath string
}

// Flags - 申明参数
func (cmd *GenerateUnitTestCommand) Flags(fs *flag.FlagSet) *flag.FlagSet {
	fs.StringVar(&cmd.projectPath, "projectPath", "", "the project path")
	return cmd.baseCommand.Flags(fs)
}

// Run - 生成代码
func (cmd *GenerateUnitTestCommand) Run(args []string) error {
	return cmd.run(args, cmd.genrateJS)
}

func (cmd *GenerateUnitTestCommand) genrateJS(cls *types.ClassSpec) error {
	ctlName := Pluralize(cls.Name)
	params := map[string]interface{}{"namespace": cmd.ns,
		"controllerName": ctlName,
		"modelName":      ctlName,
		"projectPath":    cmd.projectPath,
		"class":          cls}
	funcs := template.FuncMap{
		"omitempty": func(t types.FieldSpec) bool {
			return !t.IsRequired
		},
		"isID": func(f types.FieldSpec) bool {
			if f.Name == "id" {
				return true
			}
			return false
		},
		"tableName":   getTableName,
		"randomValue": randomValue}

	err := cmd.executeTempate(cmd.override, []string{"tests/test_ctl"}, funcs, params,
		filepath.Join(cmd.output, "tests", Underscore(Pluralize(cls.Name))+".go"))
	if err != nil {
		return errors.New("gen unittest: " + err.Error())
	}

	err = cmd.executeTempate(cmd.override, []string{"tests/test_yaml"}, funcs, params,
		filepath.Join(cmd.output, "tests", "fixtures", Underscore(Pluralize(cls.Name))+".yaml"))
	if err != nil {
		return errors.New("gen unittest: " + err.Error())
	}
	return nil
}

func randomValue(t types.FieldSpec) string {
	switch strings.ToLower(t.Type) {
	case "string", "password":
		if t.Restrictions == nil {
			return Random.Faker().Sentence(3, false)
		}
		if t.Restrictions.Length > 0 {
			return Random.Faker().Characters(t.Restrictions.Length)
		}
		if t.Restrictions.MinLength > 0 {
			//if t.Restrictions.MaxLength > 0 {
			//	return Random.Faker().Characters(t.Restrictions.MinLength + 1)
			//}
			return Random.Faker().Characters(t.Restrictions.MinLength + 1)
		}
		return Random.Faker().Sentence(3, false)
	case "integer", "biginteger":
		if t.Restrictions == nil {
			return fmt.Sprint(Random.Int("", ""))
		}
		return fmt.Sprint(Random.Int(t.Restrictions.MinValue, t.Restrictions.MaxValue))
	case "number":
		if t.Restrictions == nil {
			return fmt.Sprint(Random.Float64("", ""))
		}
		return fmt.Sprint(Random.Float64(t.Restrictions.MinValue, t.Restrictions.MaxValue))
	case "ipaddress":
		return Random.Faker().IPv4Address().String()
	case "email":
		return Random.Faker().Email()
	case "datetime":
		return Random.DateTime().Format(time.RFC3339Nano)
	default:
		return "abc"
	}
}

var Random = NewRandomGenerator(fmt.Sprint(time.Now().Unix()))

// RandomGenerator generates consistent random values of different types given a seed.
// The random values are consistent in that given the same seed the same random values get
// generated.
type RandomGenerator struct {
	Seed  string
	faker *faker.Faker
	rand  *rand.Rand
}

// NewRandomGenerator returns a random value generator seeded from the given string value.
func NewRandomGenerator(seed string) *RandomGenerator {
	hasher := md5.New()
	hasher.Write([]byte(seed))
	sint := int64(binary.BigEndian.Uint64(hasher.Sum(nil)))
	source := rand.NewSource(sint)
	ran := rand.New(source)
	faker := &faker.Faker{
		Language: "end",
		Dict:     faker.Dict["en"],
		Rand:     ran,
	}
	return &RandomGenerator{
		Seed:  seed,
		faker: faker,
		rand:  ran,
	}
}

// String produces a random string.
func (r *RandomGenerator) Faker() *faker.Faker {
	return r.faker

}

// DateTime produces a random date.
func (r *RandomGenerator) DateTime() time.Time {
	// Use a constant max value to make sure the same pseudo random
	// values get generated for a given API.
	max := time.Date(2016, time.July, 11, 23, 0, 0, 0, time.UTC).Unix()
	unix := r.rand.Int63n(max)
	return time.Unix(unix, 0)
}

// UUID produces a random UUID.
func (r *RandomGenerator) UUID() uuid.UUID {
	return uuid.NewV4()
}

// Bool produces a random boolean.
func (r *RandomGenerator) Bool() bool {
	return r.rand.Int()%2 == 0
}

// Float64 produces a random float64 value.
func (r *RandomGenerator) Float64(min, max string) float64 {
	if min == "" && max == "" {
		return r.rand.Float64()
	}

	if min == "" {
		maxValue, err := strconv.ParseFloat(max, 0)
		if err != nil {
			return r.rand.Float64()
		}

		for {
			value := r.rand.Float64()
			if value <= maxValue {
				return value
			}
		}
	}

	minValue, err := strconv.ParseFloat(min, 0)
	if err != nil {
		return r.rand.Float64()
	}
	if max == "" {
		for {
			value := r.rand.Float64()
			if value >= minValue {
				return value
			}
		}
	}

	maxValue, err := strconv.ParseFloat(max, 0)
	if err != nil {
		return r.rand.Float64()
	}
	for {
		value := r.rand.Float64()
		if value >= minValue && value <= maxValue {
			return value
		}
	}
}

func (r *RandomGenerator) Int(min, max string) int64 {
	if min == "" && max == "" {
		return r.rand.Int63()
	}

	if min == "" {
		maxValue, err := strconv.ParseInt(max, 10, 0)
		if err != nil {
			return r.rand.Int63()
		}

		for {
			value := r.rand.Int63()
			if value <= maxValue {
				return value
			}
		}
	}

	minValue, err := strconv.ParseInt(min, 10, 0)
	if err != nil {
		return r.rand.Int63()
	}
	if max == "" {
		for {
			value := r.rand.Int63()
			if value >= minValue {
				return value
			}
		}
	}

	maxValue, err := strconv.ParseInt(max, 10, 0)
	if err != nil {
		return r.rand.Int63()
	}
	for {
		value := r.rand.Int63()
		if value >= minValue && value <= maxValue {
			return value
		}
	}
}
