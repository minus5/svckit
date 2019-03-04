package gen

import "text/template"

var valueCopyMergeTemplate = template.Must(template.New("").Parse(`

func (o {{.Type}}) copy() {{.Type}} {
{{- range .Maps}}
  	m{{.Field}} := make(map[{{.Key}}]{{.Value}})
		for k, v := range o.{{.Field}} {
      m{{.Field}}[k] = v.copy()
    }
    o.{{.Field}} = m{{.Field}}
{{- end}}
  return o
}

{{- range .Maps}}
func (s {{.Field}}) merge(s2 {{.Field}}) {{.Field}} {
	for k, v2 := range s2 {
		v, f := s[k]
		if !f {
			s[k] = v2
			continue
		}
		s[k] = v.merge(v2)
	}
	return s
}
{{- end}}

func (o {{.Type}}) merge(o2 {{.Type}}) {{.Type}} {
{{- range .Maps}}
    o.{{.Field}} = o.{{.Field}}.merge(o2.{{.Field}})
{{- end}}
  return o
}
`))
