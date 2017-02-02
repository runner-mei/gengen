
type {{.class.Name}} struct { {{range $x := .class.Fields }}
  {{goify $x.Name  true}} {{gotype $x.Type}} `json:"{{underscore $x.Name}},{{omitempty $x}}"`
{{end}} }

