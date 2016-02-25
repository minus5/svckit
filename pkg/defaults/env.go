package defaults

import (
	"fmt"
	"os"
	"path"
	"strings"
)

var (
	appName  string
	hostname string
	dc       string
)

// environment varijable koje se citaju u ovom paketu
const (
	EnvNode   = "node"
	EnvNodeIp = "node_ip"
	EnvDc     = "dc"
)

func init() {
	readEnv()
}

func readEnv() {
	appName = path.Base(os.Args[0])

	hostname, _ = os.Hostname()
	if strings.Contains(hostname, ".") {
		hostname = strings.Split(hostname, ".")[0]
	}

	dc = fmt.Sprintf("dc-%s", hostname)
	if e, ok := os.LookupEnv(EnvDc); ok {
		dc = e
	}
}

func AppName() string {
	return appName
}

func Hostname() string {
	return hostname
}

func Dc() string {
	return dc
}
