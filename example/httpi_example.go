// +build httpi_example

package main

import (
	"net/http"

	"github.com/mnu5/svckit/health"
	"github.com/mnu5/svckit/httpi"
)

/* Navigiraj na:
http://localhost:8123/ping         (curl -v pa pogledaj header Application)
http://localhost:8123/health_check
http://localhost:8123/debug/vars   (pogledaj svckit.stats key i kako je implementirano)
http://localhost:8123/debug/pprof
*/
func main() {
	health.Set(func() (health.Status, []byte) {
		return health.Passing, []byte("Ok")
	})
	httpi.Route("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})
	httpi.Start(":8123")
}
