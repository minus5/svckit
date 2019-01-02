package nsqm

import (
	"net"
	"os"

	m "github.com/minus5/nsqm"
	"github.com/minus5/nsqm/discovery/consul"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"
)

const (
	// EnvConsul is location of the consul to use. If not defined local consul is used.
	EnvConsul = "SVCKIT_DCY_CONSUL"
)

// Config konfigurira nsqm paket na nas nacin
func Config() *m.Config {
	consulAddr := "127.0.0.1:8500"
	if e, ok := os.LookupEnv(EnvConsul); ok && e != "" {
		consulAddr = e
	}
	if _, _, err := net.SplitHostPort(consulAddr); err != nil {
		consulAddr = consulAddr + ":8500"
	}

	var cfg *m.Config
	connect := func() error {
		dcy, err := consul.New(consulAddr)
		if err != nil {
			return err
		}
		cfg, err = m.WithDiscovery(dcy)
		if err != nil {
			return err
		}
		return nil
	}

	if err := signal.WithExponentialBackoff(connect); err != nil {
		log.S("consulAddr", consulAddr).Fatal(err)
	}
	return cfg
}
