package main

import (
	"fmt"

	"github.com/mnu5/svckit/example/nsq_rr/api"
	"github.com/mnu5/svckit/example/nsq_rr/api/nsq"
	"github.com/mnu5/svckit/log"
)

var c *api.Client

func main() {
	log.Discard()
	c = nsq.NewClient()

	call(2, 3)
	call(128, 129)

	c.Close()
}

func call(x, y int) {
	z, err := c.Add(x, y)
	if err == api.ErrOverflow {
		fmt.Printf("%d + %d = owerflow\n", x, y)
		return
	}
	if err != nil {
		panic(err)
	}
	fmt.Printf("%d + %d = %d\n", x, y, z)
}
