package mssql

import (
	"bytes"
	"net/url"
	"text/template"

	"github.com/minus5/svckit/asm"
	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/log"
)

type Params struct {
	Username         string `json:"username"`
	Password         string `json:"password"`
	ConnectionString string `json:"connectionString"`
}

func DefaultConnStr(name string) string {
	if kvs, err := fetchKV("mssql/" + name); err == nil && kvs != nil {
		if connStr := kvs["connectionString"]; connStr != "" {
			return connStr
		}
		if dcs, err := dcy.KV("mssql/default/connectionString"); err == nil && dcs != "" {
			return connectionStringFromTemplate(dcs, kvs)
		}
	}
	return ""
}

func fetchKV(name string) (map[string]string, error) {
	kvs := map[string]interface{}{}
	err := asm.ParseKV(name, &kvs)
	log.S("name", name).I("len", len(kvs)).Info("ASM fetched")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if len(kvs) > 0 {
		ret := map[string]string{}
		for k, v := range kvs {
			switch s := v.(type) {
			case string:
				ret[k] = url.QueryEscape(s)
			}
		}
		return ret, nil
	}
	return dcy.KVs(name)
}

func connectionStringFromTemplate(tpl string, param interface{}) string {
	buf := bytes.NewBuffer(nil)
	pt := template.Must(template.New("").Parse(tpl))
	if err := pt.Execute(buf, param); err != nil {
		log.Error(err)
		return ""
	}
	return buf.String()
}
