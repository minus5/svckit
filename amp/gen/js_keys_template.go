package gen

import "text/template"

var jsKeysTemplate = template.Must(template.New("").Funcs(fns).Parse(`
  function _{{.Type}}(o) {
    var keys = {
    {{- range .Fields }}
      {{- if .NeedUnpack }}
      "{{.TagName}}": "{{.NameLower}}",
    {{- end}}
    {{- end}}
    {{- range .Maps}}
    {{- if .NeedUnpack }}
      "{{.TagName}}": "{{.FieldLower}}",
    {{- end}}
    {{- end}}

    {{- range .Structs}}
    {{- if .NeedUnpack }}
    if (o["{{.TagName}}"] !== undefined) {
      "{{.TagName}}": "{{.FieldLower}}",
    }
    {{- end}}
    {{- end}}
    };
    unpackObject(o, keys);

    {{- range .Maps}}
    unpackMap(o["{{.FieldLower}}"], _{{.Value}});
    {{- end}}
    {{- range .Structs}}
    var s = o["{{.FieldLower}}"];
    if (s !== undefined && Object.keys(s).length > 0) {
      s["_isStruct"] = true;
      _{{.Field}}(s);
    }
    {{- end}}
  }
`))
