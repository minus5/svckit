package gen

import (
	"strings"
	"text/template"
)

var fns = template.FuncMap{
	"join": strings.Join,
}

var diffStructTemplate = template.Must(template.New("").Funcs(fns).Parse(`

type {{.Type}}Diff struct {
{{- range .Fields }}
	{{.Name}} *{{.Type}} {{.Tag}}
{{- end}}
{{- range .Structs}}
	{{.Field}} *{{.Type}}Diff {{.Tag}}
{{- end}}
{{- range .Maps}}
	{{.Field}}  map[{{.Key}}]*{{.Value}}Diff {{.Tag}}
{{- end}}
}

func (o {{.Type}}Diff) empty() bool {
  return {{join .EmptyParts " &&\n"}}
}

`))
