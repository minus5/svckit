package gen

import "text/template"

var jsKeysTemplate = template.Must(template.New("").Funcs(fns).Parse(`
  function _unpack{{.Type}}(o) {
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
    _unpack(o, keys);

    {{- range .Maps}}
    for(var k in o["{{.FieldLower}}"]) {
      var c = o["{{.FieldLower}}"][k];
      if (c !== null) {
        _unpack{{.Value}}(c);
      }
    }
    {{- end}}
    {{- range .Structs}}
    var s = o["{{.FieldLower}}"];
    if (s !== undefined && Object.keys(s).length > 0) {
      s["_isStruct"] = 1;
      _unpack{{.Field}}(s);
    }
    {{- end}}
  }
`))
