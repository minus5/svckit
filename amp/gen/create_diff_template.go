package gen

import "text/template"

var createDiffTemplate = template.Must(template.New("").Parse(`

// Diff creates diff (i) between new (n) and old (o) {{.Type}}.
// So that diff applyed to old will produce new.
func (o *{{.Type}}) Diff(n *{{.Type}}) *{{.Type}}Diff {
{{- if .HasTsMethod }}
	if ot := o.Ts(); ot != 0 && ot == n.Ts() {
		return nil
	}
{{- end}}
	i := &{{.Type}}Diff{}

{{- range .Fields }}
  {{- if .IsStruct}}
	  if !n.{{.Name}}.Equal(o.{{.Name}}) {
  {{- else if .IsSlice}}
    	if !reflect.DeepEqual(n.{{.Name}}, o.{{.Name}}) {
  {{- else}}
  	if n.{{.Name}} != o.{{.Name}} {
  {{- end}}
		i.{{.Name}} = &n.{{.Name}}
	}
{{- end}}

{{- range .Structs}}
  i.{{.Field}} = o.{{.Field}}.Diff(&n.{{.Field}})
{{- end}}
{{- range .Maps}}
	i.{{.Field}} = make(map[{{.Key}}]*{{.Value}}Diff)
	for k, nc := range n.{{.Field}} {
		oc, ok := o.{{.Field}}[k]
		if !ok {
      oc = &{{.Value}}{}
		}
		ip := oc.Diff(nc)
		if ip != nil {
			i.{{.Field}}[k] = ip
		}
	}
{{- if .HasHiddenAttr }}
  for k, oc := range o.{{.Field}} {
		if _, ok := n.{{.Field}}[k]; !ok && !oc.hidden {
		  i.{{.Field}}[k] = nil // signal hide
		}
  }
{{- else }}
	for k, _ := range o.{{.Field}} {
		if _, ok := n.{{.Field}}[k]; !ok {
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
