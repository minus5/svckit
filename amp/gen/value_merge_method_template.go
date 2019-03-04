package gen

import "text/template"

var valueMergeMethodTemplate = template.Must(template.New("").Parse(`

// MergeDiff applies diff (d) to {{.Type}} (o).
func (o {{.Type}}) MergeDiff(d *{{.Type}}Diff) ({{.Type}}, bool) {
  if d == nil {
    return o, false
  }
  changed := false
// fields
{{- range .Fields }}
  {{- if .IsStruct}}
	  if d.{{.Name}} != nil && !o.{{.Name}}.Equal(*d.{{.Name}}) {
  {{- else if .IsSlice}}
  	if d.{{.Name}} != nil {
  {{- else}}
  	if d.{{.Name}} != nil && *d.{{.Name}} != o.{{.Name}} {
  {{- end}}
		o.{{.Name}} = *d.{{.Name}}
    changed = true
	}
{{- end}}

{{- range .Structs}}
  // {{.Field}} field
  if o2, merged := o.{{.Field}}.MergeDiff(d.{{.Field}}); merged {
    o.{{.Field}} = o2
    changed = true
  }
{{- end}}

{{- range .Maps}}
// {{.Field}} map
  	var copy{{.Field}}Once sync.Once
  	copyOnWrite{{.Field}} := func() {
  		copy{{.Field}}Once.Do(func() {
  			m := make(map[{{.Key}}]{{.Value}})
  			for k, v := range o.{{.Field}} {
  				m[k] = v
  			}
  			o.{{.Field}} = m
  			changed = true
  		})
  	}
		for k, dc := range d.{{.Field}} {
			c, ok := o.{{.Field}}[k]
			if dc == nil {
				if ok {
          copyOnWrite{{.Field}}()
          delete(o.{{.Field}}, k)
				}
				continue
			}
  		if c2, merged := c.MergeDiff(dc); merged {
    		copyOnWrite{{.Field}}()
  	  	o.{{.Field}}[k] = c2
      }
		}
{{- end}}
  return o, changed
}
`))
