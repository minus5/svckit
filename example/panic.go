package main

import "github.com/mnu5/svckit/log"

func main() {
	log.I("key", 1).Debug("Pero")
	var n *neki
	log.Printf("%d", n.i)
}

type neki struct {
	i *int
}
