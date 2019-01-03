package gen

import (
	"text/template"
)

var langStructTemplate = template.Must(template.New("").Funcs(fns).Parse(`

type {{.Type}}Client struct {
{{- range .Fields }}
	{{.Name}} {{.ClientType}} {{.Tag}}
{{- end}}
{{- range .Maps}}
	{{.Field}} map[{{.Key}}]*{{.Value}}Client {{.Tag}}
{{- end}}
{{- if .HasParent}}
  {{.ParentName}} *{{.ParentType}}Client
  key string
{{- end}}
}

func (o *{{.Type}}) Translate(lang string) *{{.Type}}Client {
  l := &{{.Type}}Client{}
{{- range .Fields }}
  {{- if .IsClientType}}
	l.{{.Name}} = o.{{.Name}}.Lang(lang)
  {{- else}}
	l.{{.Name}} = o.{{.Name}}
  {{- end}}
{{- end}}
{{- range .Maps}}
  l.{{.Field}} = make(map[{{.Key}}]*{{.Value}}Client)
  for k,v := range o.{{.Field}} {
   l.{{.Field}}[k] = v.Translate(lang)
  }
{{- end}}
  return l
}

type {{.Type}}ClientDiff struct {
{{- range .Fields }}
	{{.Name}} *{{.ClientType}} {{.Tag}}
{{- end}}
{{- range .Maps}}
	{{.Field}} map[{{.Key}}]*{{.Value}}ClientDiff {{.Tag}}
{{- end}}
}

func (o *{{.Type}}Diff) Translate(lang string) *{{.Type}}ClientDiff {
  l := &{{.Type}}ClientDiff{}
{{- range .Fields }}
  {{- if .IsClientType}}
  if  o.{{.Name}} != nil {
    if v := o.{{.Name}}.Lang(lang); v != "" {
  	  l.{{.Name}} = &v
    }
  }
  {{- else}}
	l.{{.Name}} = o.{{.Name}}
  {{- end}}
{{- end}}
{{- range .Maps}}
  l.{{.Field}} = make(map[{{.Key}}]*{{.Value}}ClientDiff)
  for k,v := range o.{{.Field}} {
   if v != nil {
     l.{{.Field}}[k] = v.Translate(lang)
   } else {
     l.{{.Field}}[k] = nil
   }
  }
{{- end}}
  return l
}


func (o {{.Type}}ClientDiff) empty() bool {
  return {{join .EmptyParts " &&\n"}}
}

`))
