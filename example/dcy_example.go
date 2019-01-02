// +build dcy_example

package main

import (
	"fmt"
	"github.com/mnu5/svckit/dcy"
)

func main() {
	fmt.Printf("node: %v\n", dcy.NodeName())
	fmt.Printf("dc: %v\n", dcy.Dc())

	l, _ := dcy.AgentService("nsqd-tcp")
	fmt.Printf("nsq-tcp local: %v\n", l)

	l, _ = dcy.AgentService("consul")
	fmt.Printf("nsq-tcp local: %v\n", l)

	for i := 0; i < 10; i++ {
		adr, _ := dcy.Service("test3")
		fmt.Printf("one test3 service: %v\n", adr)
	}

	adrs, _ := dcy.Services("test3")
	fmt.Printf("one test3 service: %v\n", adrs)

	adrs, _ = dcy.Services("test1")
	fmt.Printf("all test1 services: %v\n", adrs)

}
