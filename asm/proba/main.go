package main

import (
	"flag"
	"fmt"
	"github.com/minus5/svckit/asm"
	"github.com/minus5/svckit/log"
)

var name = ""

func init() {
	flag.StringVar(&name, "name", "", "name of the secret to retrieve")
	flag.Parse()
	if name == "" {
		log.Fatal(fmt.Errorf("please provide name of the secret to retreive"))
	}
}
func main() {
	ss, err := asm.GetSecretString(name)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(ss)
}
