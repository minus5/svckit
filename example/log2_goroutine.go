// +build test

package main

//
//func f(broj int) {
//	for i := 0; i < 2; i++ {
//		fmt.Println("u petlji %d , %d", broj, i)
//		log2.Info("funkcija %d, poziv %d", broj, i)
//	}
//}
//
//func main() {
//	fmt.Println("pozvana f-ja")
//	for j := 0; j < 5; j++ {
//		fmt.Println("u petlji")
//		go f(j)
//	}
//	fmt.Println("kraj")
//}

import (
	"strconv"
	"sync"
	"time"

	"github.com/minus5/svckit/log2"
)

var wg sync.WaitGroup

func f(s string) {
	l := log2.S("rutina", s).New()
	for i := 0; i < 5; i++ {
		time.Sleep(100 * time.Millisecond)
		l.I("poziv", i).Info("msg")
	}
	wg.Done()
}

func main() {
	for j := 0; j < 5; j++ {
		wg.Add(1)
		go f(strconv.Itoa(j))
	}
	//go f("world")
	//f("hello")
	wg.Wait()
}
