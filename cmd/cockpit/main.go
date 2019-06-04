package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"

	yaml "gopkg.in/yaml.v2"
)

var configFile string
var noClear bool
var servicesFileName = "services.yml"
var bindInterface = ""
var bindIP = "127.0.0.1"

func init() {
	flag.StringVar(&configFile, "config", "./cockpit.yml", "config file name")
	flag.StringVar(&bindInterface, "if", "lo0", "bind to this interface")
	flag.BoolVar(&noClear, "no-clear", false, "do not remove tmp directory")
	flag.Parse()
}

func logFilePath(name string) string {
	return fmt.Sprintf("./log/%s.log", name)
}

func interfaceIP(ifc string) string {
	i, err := net.InterfaceByName(ifc)
	if err != nil {
		log.Fatal(err)
	}
	addrs, err := i.Addrs()
	if err != nil {
		log.Fatal(err)
	}
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip.To4() != nil {
			return ip.String()
		}
	}
	return "127.0.0.1"
}

func main() {
	bindIP = interfaceIP(bindInterface)

	if fileNotExists(configFile) {
		log.Fatal(fmt.Errorf("config file [%s]is missing", configFile))
	}

	// dir := env.BinDir()
	// if err := os.Chdir(dir); err != nil {
	// 	log.Fatal(err)
	// }

	if !noClear {
		os.RemoveAll("./log")
		os.RemoveAll("./tmp")
	}

	os.MkdirAll("./log", os.ModePerm)
	os.MkdirAll("./tmp", os.ModePerm)
	os.MkdirAll("./tmp/nsqd", os.ModePerm)
	os.MkdirAll("./tmp/mongo", os.ModePerm)

	// redirect output to file
	f, err := os.Create(logFilePath(env.AppName()))
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(f)
	defer f.Close()

	log.S("bindIP", bindIP).Debug("binding")
	services := loadServices()
	config := loadConfig()
	config.services = services

	// PP(services)
	// PP(config)
	// return

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	err = config.start()
	if err == nil {
		config.startHTTP()
		<-c
	}
	fmt.Println()
	config.stop()
}

func loadServices() map[string]*service {
	files := []string{
		env.BinDir() + "/" + servicesFileName, // binary directory
		"./" + servicesFileName,               // current directory
	}
	services := make(map[string]*service)
	for _, file := range files {
		if fileNotExists(file) {
			log.S("path", file).Debug("file does not exists")
			continue
		}
		log.S("path", file).Info("loading services from file")
		for k, s := range loadServicesFile(file) {
			log.S("path", file).S("service", k).Info("service")
			services[k] = s
		}
	}
	return services
}

func fileNotExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return true
	}
	return false
}

func loadServicesFile(file string) map[string]*service {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}
	services := make(map[string]*service)
	err = yaml.Unmarshal([]byte(data), &services)
	if err != nil {
		log.Fatal(err)
	}
	for name, s := range services {
		s.init(name)
	}
	return services
}

func loadConfig() config {
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	c := config{}
	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		log.Fatal(err)
	}
	log.S("path", configFile).I("services", len(c.Services)).Debug("config")
	return c
}

// PP prety print object
func PP(o interface{}) {
	buf, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("pp:\n%s\n", buf)
}
