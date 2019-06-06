package env

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	envNode = "node"
)

var (
	dc       string
	nodeName string
	appName  string
	hostname string
)

func init() {
	appName = path.Base(os.Args[0])
	if appName == "main" { // when running with go run, get directory name instead of main
		wd, err := os.Getwd()
		if err == nil {
			_, appName = path.Split(wd)
		}
	}
	if job, ok := os.LookupEnv("NOMAD_JOB_NAME"); ok {
		appName = job
	}

	hostname, _ = os.Hostname()
	if strings.Contains(hostname, ".") {
		hostname = strings.Split(hostname, ".")[0]
	}

	if node := os.Getenv(envNode); node != "" {
		nodeName = node
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

func InDev() bool {
	return dc == "dev"
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

// Port gets port for label from evn variables
func Port(name string) int {
	if name == "" {
		name = "default"
	}
	for _, v := range []string{"NOMAD_PORT_", "PORT_"} {
		env, ok := os.LookupEnv(v + name)
		if ok {
			if p, err := strconv.Atoi(env); err == nil {
				return p
			}
		}
	}
	// TODO mybe return random port
	return 0
}

// Address gets addres for label
func Address(label string) string {
	return fmt.Sprintf(":%d", Port(label))
}

func Deployment() string {
	dep, ok := os.LookupEnv("deployment")
	if !ok {
		return dc
	}
	return dep
}
