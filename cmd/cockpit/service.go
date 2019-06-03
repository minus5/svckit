package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/hashicorp/consul/api"
	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"
)

var (
	info  = color.New(color.FgCyan).PrintfFunc()
	info2 = color.New(color.FgMagenta).PrintfFunc()
	warn  = color.New(color.FgRed).Add(color.Bold).PrintfFunc()
)

type service struct {
	Name       string
	Entrypoint string
	Command    string
	Path       string
	WatchPath  bool `yaml:"watch_path"`
	Build      string
	Topics     []string
	done       chan struct{}
	cmd        *exec.Cmd
	Port       int
	Consul     []*serviceConsul
	watcher    *fsnotify.Watcher
	Kill       string
	Env        []string
	KV         map[string]string
}

type serviceConsul struct {
	Name      string
	PortLabel string `yaml:"port_label"`
	Port      int
	Tags      []string
	Address   string
	HTTPCheck string `yaml:"http_check"`
}

var netPortRange = 9000

func netPort() int {
	netPortRange++
	return netPortRange
}

func (s *service) init(name string) {
	if s.Path != "" {
		s.Path = env.ExpandPath(s.Path)
	}
	s.Name = name
	if s.Entrypoint == "" {
		s.Entrypoint = name
	}
	if strings.Contains(s.Command, "$USER") {
		s.Command = strings.Replace(s.Command, "$USER", env.Username(), -1)
	}
	if s.Port != 0 {
		s.Consul = append(s.Consul, &serviceConsul{
			Port:      s.Port,
			Name:      name,
			HTTPCheck: "/health_check",
		})
	}
	for _, c := range s.Consul {
		// if c.HTTPCheck == "" {
		// 	c.HTTPCheck = "/health_check"
		// }
		if c.Name == "" {
			c.Name = name
		}
		if c.Port == 0 {
			c.Port = netPort()
		}
		s.Env = append(s.Env, fmt.Sprintf("PORT_%s=%d", c.PortLabel, c.Port))
	}
}

func (s service) String() string {
	return s.Name
}

func (s service) logFile() (*os.File, error) {
	return os.Create(logFilePath(s.Name))
}

func (s *service) stop() {
	if s == nil || s.cmd == nil || s.done == nil {
		return
	}
	select {
	case <-s.done:
		return
	default:
	}

	if s.Kill == "hup" {
		terminateProc(s.cmd.Process)
	} else {
		s.cmd.Process.Signal(os.Interrupt)
	}
	select {
	case <-s.done:
		return
	case <-time.After(20 * time.Second):
		s.cmd.Process.Signal(os.Kill)
		select {
		case <-s.done:
			return
		case <-time.After(10 * time.Second):
			warn("Failed to stop %s\n", s)
		}
	}
}

// ukradeno iz goreman-a
// ref: https://github.com/mattn/goreman/blob/d0ee41b21be92ce6fd3e55ad11c5e5c9452fe822/proc_posix.go#L43
func terminateProc(p *os.Process) error {
	pgid, err := syscall.Getpgid(p.Pid)
	if err != nil {
		log.Error(err)
		return err
	}

	// use pgid, ref: http://unix.stackexchange.com/questions/14815/process-descendants
	pid := p.Pid
	if pgid == p.Pid {
		pid = -1 * pid
	}

	target, err := os.FindProcess(pid)
	if err != nil {
		log.Error(err)
		return err
	}
	err = target.Signal(syscall.SIGHUP)
	if err != nil {
		log.Error(err)
	}
	return err
}

func fileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func (s *service) entrypoint() string {
	//cs := p.cmdLine() //[]string{"/bin/sh", "-c", "exec " + p.cmd + p.logLine()}
	//cmd := exec.Command(cs[0], cs[1:]...)
	// fmt.Println("exec", s.Entrypoint, s.Command)
	e := s.Entrypoint
	if s.Path != "" &&
		!strings.HasPrefix(s.Entrypoint, "/") &&
		!strings.Contains(s.Entrypoint, "=") &&
		fileExists(s.Path+"/"+e) {
		e = "./" + e
	}
	return e
}

func (s *service) start() error {
	if s.Entrypoint == "_" || strings.HasSuffix(s.Name, "_build") {
		info("Done %s\n", s)
		return nil
	}
	logFile, err := s.logFile()
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd := exec.Command(s.entrypoint(), strings.Split(s.Command, " ")...)
	if len(s.Env) != 0 {
		path := os.Getenv("PATH")
		s.Env = append(s.Env, "PATH="+path)
		cmd.Env = s.Env
	}
	cmd.Stdin = nil
	if s.Path != "" {
		cmd.Dir = s.Path
	}
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		warn("Failed to start %s: %s\n", s, err)
		return err
	}
	info("Started %s\n", s)

	s.cmd = cmd
	done := make(chan struct{})
	s.done = done

	go func() {
		defer close(done)
		err = cmd.Wait()
		if err != nil {
			info2("Stopped %s, %s\n", s, err)
			return
		}
		info2("Stopped %s\n", s)
	}()

	return s.Watch()
}

func (s *service) register() error {
	for _, c := range s.Consul {
		if c.Name == "" {
			c.Name = s.Name
		}
		if c.PortLabel != "" {
			c.Tags = append(c.Tags, c.PortLabel)
		}
		r := func() error {
			return register(c)
		}
		if err := signal.WithExponentialBackoff(r); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}

func (s *service) kv() error {
	config := api.DefaultConfig()
	consul, err := api.NewClient(config)
	if err != nil {
		log.Error(err)
		return err
	}
	for k, v := range s.KV {
		p := api.KVPair{
			Key:   k,
			Value: []byte(v),
		}
		_, err := consul.KV().Put(&p, nil)
		if err != nil {
			return err
		}
		log.S("key", k).S("value", v).Debug("kv")
	}
	return nil
}

func (s *service) prepare() error {
	if s.Build == "" {
		return nil
	}
	c := strings.Split(s.Build, " ")
	cmd := exec.Command(c[0], c[1:]...)
	if s.Path != "" {
		cmd.Dir = s.Path
	}
	if err := cmd.Start(); err != nil {
		log.Error(err)
		return err
	}
	return cmd.Wait()
}

func (s *service) createTopics() error {
	for _, topic := range s.Topics {
		c := func() error { return createTopic(topic) }
		if err := signal.WithExponentialBackoff(c); err != nil {
			return err
		}
	}
	return nil
}

func (s *service) Go() error {
	if err := s.prepare(); err != nil {
		return err
	}
	if err := s.createTopics(); err != nil {
		return err
	}
	if err := s.register(); err != nil {
		return err
	}
	if err := s.kv(); err != nil {
		return err
	}
	if err := s.start(); err != nil {
		return err
	}
	return nil
}

func register(c *serviceConsul) error {
	config := api.DefaultConfig()
	consul, err := api.NewClient(config)
	if err != nil {
		log.Error(err)
		return err
	}

	agent := consul.Agent()

	checkID := c.Name
	if c.PortLabel != "" {
		checkID = checkID + "-" + c.PortLabel
	}

	service := &api.AgentServiceRegistration{
		ID:      checkID,
		Name:    c.Name,
		Port:    c.Port,
		Tags:    c.Tags,
		Address: c.Address,
	}
	if err := agent.ServiceRegister(service); err != nil {
		log.Error(err)
		return err
	}
	log.S("service", c.Name).I("port", c.Port).Info("registerd service")

	if c.HTTPCheck == "" {
		tcp := fmt.Sprintf("127.0.0.1:%d", c.Port)
		if c.Address != "" {
			tcp = fmt.Sprintf("%s:%d", c.Address, c.Port)
		}
		check := &api.AgentCheckRegistration{
			ID:        checkID,
			Name:      fmt.Sprintf("Service '%s' TCP health check", checkID),
			Notes:     fmt.Sprintf("tcp: %s", tcp),
			ServiceID: service.ID,
			AgentServiceCheck: api.AgentServiceCheck{
				Status:   api.HealthPassing,
				Interval: "10s",
				Timeout:  "1s",
				TCP:      tcp,
			},
		}
		if err := agent.CheckRegister(check); err != nil {
			log.Error(err)
			return err
		}
		log.S("service", c.Name).S("checkID", checkID).S("tcp", tcp).Info("registerd health check")
		return nil
	}

	url := c.HTTPCheck
	if !strings.HasPrefix(url, "http") {
		url = fmt.Sprintf("http://127.0.0.1:%d%s", c.Port, c.HTTPCheck)
	}
	check := &api.AgentCheckRegistration{
		ID:        checkID,
		Name:      fmt.Sprintf("Service '%s' HTTP health check", checkID),
		Notes:     fmt.Sprintf("url: %s", url),
		ServiceID: service.ID,
		AgentServiceCheck: api.AgentServiceCheck{
			Status:   api.HealthPassing,
			Interval: "10s",
			Timeout:  "1s",
			HTTP:     url,
		},
	}
	if err := agent.CheckRegister(check); err != nil {
		log.Error(err)
		return err
	}
	log.S("service", c.Name).S("checkID", checkID).S("url", url).Info("registerd health check")

	return nil
}

func (s *service) Watch() error {
	if s.Path == "" {
		return nil
	}
	if s.watcher != nil {
		return nil
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
		return err
	}

	go func() {
		defer watcher.Close()

		for {
			select {
			case event := <-watcher.Events:
				//log.S("service", s.Name).Debug(fmt.Sprintf("event: %#v", event))
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create {
					log.S("service", s.Name).S("file", event.Name).Info("modified")
					s.stop()
					s.start()
				}
			case err := <-watcher.Errors:
				if err != nil {
					log.S("service", s.Name).Error(err)
				}
			}
		}
	}()

	if bin := s.Path + "/" + s.Entrypoint; exists(bin) {
		// watch binary
		log.S("path", bin).S("service", s.Name).Info("added file watcher")
		if err = watcher.Add(bin); err != nil {
			log.Error(err)
			return err
		}
	} else {
		if !s.WatchPath {
			return nil
		}
		// watch path recursively
		err := filepath.Walk(s.Path, func(path string, f os.FileInfo, err error) error {
			if f.IsDir() && !(strings.Contains(path, "/tmp/") || strings.Contains(path, "/log/")) {
				log.S("path", path).S("service", s.Name).Info("added folder watcher")
				return watcher.Add(path)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	s.watcher = watcher
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func createTopic(topic string) error {
	if strings.HasPrefix(topic, "z...rsp.") {
		topic = topic + ".dev01"
	}
	url := fmt.Sprintf("http://127.0.0.1:4151/topic/create?topic=%s", topic)
	rsp, err := http.Post(url, "", nil)
	if err != nil {
		log.Error(err)
		return err
	}
	if rsp.StatusCode != http.StatusOK {
		err := fmt.Errorf("http status %d", rsp.StatusCode)
		log.Error(err)
		return err
	}
	log.S("topic", topic).Info("nsq topic created")
	return nil
}
