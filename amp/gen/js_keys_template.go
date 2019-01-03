package gen

import "text/template"

var jsKeysTemplate = template.Must(template.New("").Funcs(fns).Parse(`
  function unpack{{.Type}}(o) {
  {{- range .Fields }}
    {{- if .NeedUnpack }}
    if (o["{{.TagName}}"] !== undefined) {
      o["{{.NameLower}}"] = o["{{.TagName}}"];
      delete o["{{.TagName}}"];
     }
    {{- end}}
  {{- end}}
  {{- range .Maps}}
    {{- if .NeedUnpack }}
    if (o["{{.TagName}}"] != undefined) {
      o["{{.FieldLower}}"] = o["{{.TagName}}"];
      delete o["{{.TagName}}"];
     }
    {{- end}}
     for(var k in o["{{.FieldLower}}"]) {
       var c = o["{{.FieldLower}}"][k];
       if (c !== null) {
         unpack{{.Value}}(c);
       }
     }
  {{- end}}
  {{- range .Structs}}
    {{- if .NeedUnpack }}
    if (o["{{.TagName}}"] !== undefined) {
      o["{{.FieldLower}}"] = o["{{.TagName}}"];
      delete o["{{.TagName}}"];
    }
    {{- end}}
    var s = o["{{.FieldLower}}"];
    if (s !== undefined && Object.keys(s).length > 0) {
      s["_isStruct"] = 1;
      unpack{{.Field}}(s);
    }
  {{- end}}
  }
`))
