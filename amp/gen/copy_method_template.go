package gen

import "text/template"

var copyMethodTemplate = template.Must(template.New("").Parse(`

func (o *{{.Type}}) Copy() *{{.Type}} {
  o2 := &{{.Type}}{}
  *o2 = *o
{{- range .Maps}}
  	o2.{{.Field}} = make(map[{{.Key}}]*{{.Value}})
		for k, oc := range o.{{.Field}} {
      o2.{{.Field}}[k] = oc.Copy()
    }
{{- end}}
  return o2
}

// IncrementalDiff makes diff between partial full n and current o
func (o *{{.Type}}) IncrementalDiff(n *{{.Type}}) *{{.Type}}Diff {
  if n == nil {
    return nil
  }
	nd := n.ToDiff()
  if o == nil {
    return nd
  }
  // apply that diff to copy of o
	n = o.Copy()
	n.MergeDiff(nd)
  // create diff
	return o.Diff(n)
}

// ToDiff converts n to diff
func (o *{{.Type}}) ToDiff() *{{.Type}}Diff {
	empty := &{{.Type}}{}
	return empty.Diff(o)
}

`))
