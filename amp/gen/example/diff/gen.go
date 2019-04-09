// The following directive is necessary to make the package coherent:

// +build ignore

package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/minus5/svckit/amp/gen"
	"github.com/minus5/svckit/amp/gen/example/diff"
)

func main() {
	//apiPath := "../../api/js/src/v1_unpack.js"
	b := &diff.Book{}
	genDiff(b, "sport")
	//jsUnpack(b, apiPath)
}

func genDiff(o interface{}, prefix string) {
	v := reflect.ValueOf(o).Elem()
	if err := gen.ValueDiff(v, fmt.Sprintf("./%s_diff.go", prefix)); err != nil {
		log.Fatal(err)
	}
	if err := gen.ValueCopyMerge(v, fmt.Sprintf("./%s_copy_merge.go", prefix)); err != nil {
		log.Fatal(err)
	}
}

func jsUnpack(o interface{}, filename string) {
	v := reflect.ValueOf(o).Elem()
	if err := gen.JsKeys(v, filename); err != nil {
		log.Fatal(err)
	}
}
