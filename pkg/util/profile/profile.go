package profile

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"

	"github.com/minus5/svckit/log"
)

// Start pokrece cpu profile, vraca file u koji se snima
func Start() string {
	output := fmt.Sprintf("./%s.pprof", time.Now().Format(time.RFC3339))
	log.S("output", output).Info("starting profile")
	f, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		log.Fatal(err)
	}
	return output
}

// Stop zaustavlja cpu profile
func Stop() {
	pprof.StopCPUProfile()
}
