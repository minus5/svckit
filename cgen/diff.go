package cgen

import (
	"bytes"
	"log"
	"os/exec"
	"strings"
	"text/template"
)

func diff(data []Struct) string {
	return string(runTemplate(diffTemplate, data))
}

func merge(data []Struct) string {
	return string(runTemplate(mergeTemplate, data))
}

func diffMethods(data []Struct) string {
	return string(runTemplate(diffMethodsTemplate, data))
}

func runTemplate(tpl *template.Template, data interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		log.Fatal(err)
	}
	return gofmt(buf.Bytes())
}

func gofmt(in []byte) []byte {
	cmd := exec.Command("gofmt")
	cmd.Stdin = strings.NewReader(string(in))
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return out
}

var diffTemplate = template.Must(template.New("").Parse(`

{{- range . }}
type {{.Type}}Diff struct {
{{- range .Fields }}
	{{.Name}} *{{.Type}} {{.Tag}}
{{- end}}
{{- range .Structs}}
	{{.Name}} *{{.Type}}Diff {{.Tag}}
{{- end}}
{{- range .Maps}}
	{{.Field}}  map[{{.Key}}]*{{.Value}}Diff {{.Tag}}
{{- end}}
}
{{- end}}

`))

var mergeTemplate = template.Must(template.New("").Parse(`

{{- range . }}
// Merge applies diff (d) to {{.Type}} (o)
// and returns new value type with merged changes.
// Doesn't modifies original value (o).
func (o {{.Type}}) Merge(d *{{.Type}}Diff) {{.Type}} {
  n, _ := o.merge()
  return n
}

func (o {{.Type}}) merge(d *{{.Type}}Diff) ({{.Type}}, bool) {
  if d == nil {
    return o, false
  }
  changed := false
// fields
{{- range .Fields }}
  if d.{{.Name}} != nil && *d.{{.Name}} != o.{{.Name}} {
		o.{{.Name}} = *d.{{.Name}}
    changed = true
	}
{{- end}}

{{- range .Structs}}
  // {{.Name}} field
  if o2, merged := o.{{.Name}}.merge(d.{{.Name}}); merged {
    o.{{.Name}} = o2
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
  		if c2, merged := c.merge(dc); merged {
    		copyOnWrite{{.Field}}()
  	  	o.{{.Field}}[k] = c2
      }
		}
{{- end}}
  return o, changed
}
{{- end}}
`))

var fns = template.FuncMap{
	"join": strings.Join,
}

var diffMethodsTemplate = template.Must(template.New("").Funcs(fns).Parse(`

{{- range . }}
// Diff creates diff (i) between new (n) and old (o) {{.Type}}.
// So that diff applyed to old will produce new.
func (o {{.Type}}) Diff(n {{.Type}}) *{{.Type}}Diff {
	i := &{{.Type}}Diff{}

{{- range .Fields }}
  if n.{{.Name}} != o.{{.Name}} {
		i.{{.Name}} = &n.{{.Name}}
	}
{{- end}}

{{- range .Structs}}
  i.{{.Name}} = o.{{.Name}}.Diff(n.{{.Name}})
{{- end}}
{{- range .Maps}}
	i.{{.Field}} = make(map[{{.Key}}]*{{.Value}}Diff)
	for k, nc := range n.{{.Field}} {
		oc, ok := o.{{.Field}}[k]
		if !ok {
      oc = {{.Value}}{}
		}
		ip := oc.Diff(nc)
		if ip != nil {
			i.{{.Field}}[k] = ip
		}
	}

	for k, _ := range o.{{.Field}} {
		if _, ok := n.{{.Field}}[k]; !ok {
			i.{{.Field}}[k] = nil 
		}
  }

	if len(i.{{.Field}}) == 0 {
		i.{{.Field}} = nil
	}
{{- end}}
	if i.empty() {
		return nil
	}
	return i
}

func (i {{.Type}}Diff) empty() bool {
  return {{join .NilConditions " &&\n"}}
}

{{- end}}

`))
