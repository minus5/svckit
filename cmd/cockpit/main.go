package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/mnu5/svckit/env"
	"github.com/mnu5/svckit/log"

	yaml "gopkg.in/yaml.v2"
)

// Ideje:
//  - istaziti sto treba za web (Rails, Nginx...)
//  - razlicite konfiguracije za npr. cloudionica

var configFile string
var clear bool

func init() {
	flag.StringVar(&configFile, "config", "cloudionica.yml", "config file")
	flag.BoolVar(&clear, "c", false, "clear tmp directory")
	flag.Parse()
}

func logFilePath(name string) string {
	return fmt.Sprintf("./log/%s.log", name)
}

func main() {
	dir := env.BinDir()
	if err := os.Chdir(dir); err != nil {
		log.Fatal(err)
	}

	if clear {
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

	services := loadServices()
	config := loadConfig()
	config.services = services

	//PP(services)
	//PP(config)
	//return

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	err = config.Start()
	if err == nil {
		config.StartHTTP()
		<-c
	}
	fmt.Println()
	config.Stop()
}

func loadServices() map[string]*Service {
	data, err := ioutil.ReadFile("./services.yml")
	if err != nil {
		log.Fatal(err)
	}
	services := make(map[string]*Service)
	err = yaml.Unmarshal([]byte(data), &services)
	if err != nil {
		log.Fatal(err)
	}
	for name, s := range services {
		s.Init(name)
	}
	return services
}

func loadConfig() Config {
	data, err := ioutil.ReadFile("./" + configFile)
	if err != nil {
		log.Fatal(err)
	}
	c := Config{}
	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		log.Fatal(err)
	}
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
