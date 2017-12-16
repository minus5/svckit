package env

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

var (
	dc       string
	nodeName string
	appName  string
	hostname string
)

func init() {
	appName = path.Base(os.Args[0])

	hostname, _ = os.Hostname()
	if strings.Contains(hostname, ".") {
		hostname = strings.Split(hostname, ".")[0]
	}
}

func AppName() string {
	return appName
}

func Hostname() string {
	return hostname
}

func Dc() string {
	if dc != "" {
		return dc
	}
	return fmt.Sprintf("dc-%s", hostname)
}

func NodeName() string {
	if nodeName != "" {
		return nodeName
	}
	return hostname
}

func SetAppName(name string) {
	appName = name
}

func SetDc(name string) {
	dc = name
}

func SetNodeName(name string) {
	if strings.Contains(name, ".") {
		name = strings.Split(name, ".")[0]
	}
	nodeName = name
}

// Hack to know that I'm in running in tests http://stackoverflow.com/a/36666114
func InTest() bool {
	return flag.Lookup("test.v") != nil
}

func HomeDir() string {
	usr, _ := user.Current()
	return usr.HomeDir
}

func Username() string {
	usr, _ := user.Current()
	return usr.Username
}

func ExpandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		return strings.Replace(path, "~", HomeDir(), 1)
	}
	return path
}

func BinDir() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	exPath := filepath.Dir(ex)
	return exPath
}

func InstanceId() string {
	env, ok := os.LookupEnv("NOMAD_ALLOC_INDEX")
	if !ok {
		return NodeName()
	}
	return env
}
