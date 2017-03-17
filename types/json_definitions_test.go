package types

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v2"

	"testing"
)

var aClass = ClassSpec{
	Super: "ab",
	Name:  "Book",
	Keys:  [][]string{{"name"}},
	Fields: []FieldSpec{
		{
			Name:         "name",
			Description:  "name",
			Type:         "string",
			Collection:   false,
			IsEmbedded:   true,
			IsRequired:   true,
			IsReadOnly:   true,
			IsUniquely:   true,
			DefaultValue: "default",
			Unit:         "page",
			Restrictions: &RestrictionSpec{
				MinLength: "5",
				MaxLength: "250",
			},
			Annotations: []interface{}{
				1, 2, 3, "str"}},
		{
			Name:         "description",
			Description:  "abc",
			Type:         "string",
			DefaultValue: "default",
			Unit:         "page",
			Annotations: []interface{}{
				1, 2, 3, "str"}},
		{
			Name:         "authors",
			Description:  "abc",
			Type:         "string",
			Collection:   true,
			DefaultValue: "default",
			Unit:         "page",
			Annotations: []interface{}{
				1, 2, 3, "str"}},
		{
			Name:         "tags",
			Description:  "abc",
			Type:         "string",
			Collection:   true,
			DefaultValue: "default",
			Unit:         "page",
			Restrictions: &RestrictionSpec{
				Enumerations: []string{"a", "b", "c"},
			},
			Annotations: []interface{}{
				1, 2, 3, "str"}},
	},
}

func TestJSON(t *testing.T) {
	err := json.NewEncoder(os.Stdout).Encode(aClass)
	if err != nil {
		t.Error(err)
	}

	fmt.Println("==========")

	bs, err := yaml.Marshal(aClass)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(bs))
}
