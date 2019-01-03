package gen

import "text/template"

var adapterDiffTemplate = template.Must(template.New("").Parse(`

type {{.Type}}Adapter interface {
{{- range .Fields }}
	{{.Name}}() {{.Type}} 
{{- end}}
{{- range .Maps}}
	{{.Field}}()  map[{{.Key}}]{{.Value}}Adapter
{{- end}}
}

// Creates diff (i) between new (n) and old (o) {{.Type}}.
// So that diff applyed to old will produce new.
func (o *{{.Type}}) AdapterDiff(n {{.Type}}Adapter) *{{.Type}}Diff {
	i := &{{.Type}}Diff{}

{{- range .Fields }}
  {{- if .IsStruct}}
  	if v := n.{{.Name}}(); !v.Equal(o.{{.Name}}) {
  {{- else if .IsSlice}}
    if v := n.{{.Name}}(); !reflect.DeepEqual(v,  o.{{.Name}}) {
  {{- else}}
  	if v := n.{{.Name}}(); v != o.{{.Name}} {
  {{- end}}
		i.{{.Name}} = &v
	}
{{- end}}

{{- range .Maps}}
	i.{{.Field}} = make(map[{{.Key}}]*{{.Value}}Diff)
  n{{.Field}} := n.{{.Field}}()
	for k, nc := range n{{.Field}}  {
		oc, ok := o.{{.Field}}[k]
		if !ok {
      oc = &{{.Value}}{}
		}
		ic := oc.AdapterDiff(nc)
		if ic != nil {
			i.{{.Field}}[k] = ic
		}
	}
{{- if .HasHiddenAttr }}
  for k, oc := range o.{{.Field}} {
		if _, ok := n{{.Field}}[k]; !ok && !oc.hidden {
		  i.{{.Field}}[k] = nil // signal hide
		}
  }
{{- else }}
	for k, _ := range o.{{.Field}} {
		if _, ok := n{{.Field}}[k]; !ok {
			i.{{.Field}}[k] = nil // signal delete
		}
  }
{{- end}}
	if len(i.{{.Field}}) == 0 {
		i.{{.Field}} = nil
	}
{{- end}}
	if i.empty() {
		return nil
	}
	return i
}

`))
