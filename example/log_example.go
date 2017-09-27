// +build log_example

package main

import (
	"fmt"
	golog "log"

	"github.com/minus5/svckit/env"
	log "github.com/minus5/svckit/log"
)

func main() {
	defer log.Debug("stopped")

	//ovo je zamjena za postojecit logger
	log.Printf("pero zdero")
	log.Printf("[INFO] pero zdero")
	log.Printf("[NOTICE] pero zdero %d", 123)

	log.Info("neki info")
	log.Error(fmt.Errorf("neki error se pojavio u app %s", env.AppName()))
	log.Errorf("neki error")
	log.Notice("neki notice")

	golog.Printf("[NOTICE] sto bude kada u istoj app koristim classic logger")

	// Kako dodati atribute
	log.S("key1", "value2").S("key2", "value2 value2").I("keyi", 12345).Debug("neka poruka koja ide na kraju")
	// Cak i neki duzi string
	log.S("stack", `~/work/services/src/github.com/minus5/svckit/example>go run pero.go
/Users/ianic/work/services/src/github.com/minus5/svckit/example/pero.go:13 (0x21e3)
	main.func1: debug.PrintStack()
/usr/local/Cellar/go/1.5.2/libexec/src/runtime/asm_amd64.s:437 (0x598ee)
	call32: CALLFN(·call32, 32)
/usr/local/Cellar/go/1.5.2/libexec/src/runtime/panic.go:423 (0x2adf9)
	gopanic: reflectcall(nil, unsafe.Pointer(d.fn), deferArgs(d), uint32(d.siz), uint32(d.siz))
/usr/local/Cellar/go/1.5.2/libexec/src/runtime/panic.go:42 (0x294b9)
	panicmem: panic(memoryError)
/usr/local/Cellar/go/1.5.2/libexec/src/runtime/sigpanic_unix.go:24 (0x3fa49)
	sigpanic: panicmem()
/Users/ianic/work/services/src/github.com/minus5/svckit/example/pero.go:25 (0x20ec)
	main: log.Printf("%d", n.i)
/usr/local/Cellar/go/1.5.2/libexec/src/runtime/proc.go:111 (0x2d2e0)
	main: main_main()
/usr/local/Cellar/go/1.5.2/libexec/src/runtime/asm_amd64.s:1721 (0x5bc31)
	goexit: BYTE	$0x90	// NOP
goroutine 1 [running]:
main.main.func1()
	/Users/ianic/work/services/src/github.com/minus5/svckit/example/pero.go:17 +0x145
main.main()
	/Users/ianic/work/services/src/github.com/minus5/svckit/example/pero.go:25 +0xac

goroutine 17 [syscall, locked to thread]:
runtime.goexit()
	/usr/local/Cellar/go/1.5.2/libexec/src/runtime/asm_amd64.s:1721 +0x1`).Info("stack trace")

	//Primjer za context logging.
	//Ako zelim imati isti set atributa na vise mjesta, a da ih ne moram svaki put dodavati.
	//Napravim funkciju koja doda atribute.
	ctx := func() *log.Agregator {
		return log.S("common1", "jedan").I("common2", 2)
	}
	ctx().I("atrr", 1).I("attr2", 2).Debug("one")
	ctx().Info("two")

	//log.Fatal("neki fatal")
}
