// +build dcy_example

package main

import (
	"fmt"

	"github.com/minus5/svckit/dcy"
)

func main() {
	adrs, _ := dcy.MongoConnStr()
	fmt.Printf("mongoConnStr: %v\n", adrs)

	adrs, _ = dcy.MongoConnStr("mongo.service.sd")
	fmt.Printf("mongoConnStr: %v\n", adrs)

	adrs, _ = dcy.MongoConnStr("f1-mongo", "mongo")
	fmt.Printf("mongoConnStr: %v\n", adrs)
}
