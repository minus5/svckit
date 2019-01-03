package gen

import "text/template"

var mergeMethodTemplate = template.Must(template.New("").Parse(`

{{ $typeLower := .TypeLower }}

// MergeDiff applies diff (d) to {{.Type}} (o).
func (o *{{.Type}}) MergeDiff(d *{{.Type}}Diff) bool {
  if d == nil {
    return false
  }
  changed := false

{{- range .Fields }}
  {{- if .HasChangedMethod }}
  	{{.NameLower}}Changed := false
  	prev{{.Name}} := o.{{.Name}}
  {{- end}}
  {{- if .IsStruct}}
	  if d.{{.Name}} != nil && !o.{{.Name}}.Equal(*d.{{.Name}}) {
  {{- else if .IsSlice}}
  	if d.{{.Name}} != nil {
  {{- else}}
  	if d.{{.Name}} != nil && *d.{{.Name}} != o.{{.Name}} {
  {{- end}}
		o.{{.Name}} = *d.{{.Name}}
{{- if .HasChangedMethod }}
		{{.NameLower}}Changed = true
{{- end}}
    changed = true
	}
{{- end}}

{{- range .Fields }}
{{- if .HasChangedMethod }}
	if {{.NameLower}}Changed {
		o.{{.Name}}Changed(prev{{.Name}})
	}
{{- end}}
{{- end}}

{{- range .Structs}}
  if d.{{.Field}} != nil {
    if o.{{.Field}}.MergeDiff(d.{{.Field}}) {
      changed = true
    }
  }
{{- end}}

{{- range .Maps}}
		for k, dc := range d.{{.Field}} {
			c, ok := o.{{.Field}}[k]
			if dc == nil {
				if ok {
{{- if .HasHiddenAttr }}
          c.hidden = true
{{- else }}
          delete(o.{{.Field}}, k)
{{- end}}
				}
        changed = true
				continue
			}
			if !ok {
        c = &{{.Value}}{
{{- if .HasParent }}
				  {{.ParentAttr}}: o,
{{- end}}
{{- if .HasKeyAttr }}
				  key: k,
{{- end}}
				}
				if o.{{.Field}} == nil {
					o.{{.Field}} = make(map[{{.Key}}]*{{.Value}})
				}
				o.{{.Field}}[k] = c
			}
      {{- if .HasHiddenAttr }}
      if c.hidden {
        c.hidden = false
        changed = true
      }
      {{- end}}
			if c.MergeDiff(dc) {
        changed = true
      }
		}
{{- end}}
{{- if .HasChangedMethod}}
if changed {
  o.Changed()
}
{{- end}}
  return changed
}
`))
