package dcy

import (
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/minus5/svckit/env"
	"github.com/minus5/svckit/log"
	"github.com/minus5/svckit/signal"

	"github.com/hashicorp/consul/api"
)

const (
	// EnvConsul is location of the consul to use. If not defined local consul is used.
	EnvConsul = "SVCKIT_DCY_CONSUL"

	// EnvWait if defined dcy will not start until those services are not found in consul.
	// Usefull in development environment to controll start order.
	EnvWait = "SVCKIT_DCY_CHECK_SVCS"
)

const (
	queryTimeoutSeconds = 30
	queryRetries        = 10
	waitTimeMinutes     = 10
	localConsulAdr      = "127.0.0.1:8500"
)

var (
	consul *api.Client
	l      sync.RWMutex
	cache  = map[string][]Address{}

	domain        string
	dc            string
	nodeName      string
	advertiseAddr string
	bindAddr      string
	consulAddr    = localConsulAdr
)

// Address is service address returned from Consul.
type Address struct {
	Address string
	Port    int
}

// String return address in host:port string.
func (a Address) String() string {
	return fmt.Sprintf("%s:%d", a.Address, a.Port)
}

// Addresses is array of service addresses.
type Addresses []Address

// String returns string array in host:port format.
func (a Addresses) String() []string {
	addrs := []string{}
	for _, addr := range a {
		addrs = append(addrs, addr.String())
	}
	return addrs
}

// On including package it will try to find consul.
// Will BLOCK until consul is found.
// If not found it will raise fatal.
// To disable finding consul, and use it in test mode set EnvConsul to "-"
// If EnvWait is defined dcy will not start until those services are not found in consul. This is usefull for development environment where we start consul, and other applications which are using dcy.
func init() {
	if e, ok := os.LookupEnv(EnvConsul); ok && e != "" {
		consulAddr = e
	}
	if consulAddr == "-" || (env.InTest() && consulAddr == localConsulAdr) {
		noConsulTestMode()
		return
	}
	if _, _, err := net.SplitHostPort(consulAddr); err != nil {
		consulAddr = consulAddr + ":8500"
	}
	rand.Seed(time.Now().UTC().UnixNano())

	mustConnect()
	updateEnv()
}

func updateEnv() {
	if dc != "" {
		env.SetDc(dc)
	}
	if nodeName != "" {
		env.SetNodeName(nodeName)
	}
}

func noConsulTestMode() {
	//log.Info("setting dcy into test mode, no Consul connection")
	domain = "sd"
	dc = "dev"
	nodeName = "node01"
	bindAddr = "127.0.0.1"
	advertiseAddr = "127.0.0.1"
	cache["test1"] = []Address{
		{"127.0.0.1", 12345},
		{"127.0.0.1", 12348},
	}
	cache["test2"] = []Address{
		{"10.11.12.13", 1415},
	}
	cache["test3"] = []Address{
		{"192.168.0.1", 12345},
		{"10.0.13.0", 12347},
	}
	cache["syslog"] = []Address{
		{"127.0.0.1", 9514},
	}
	cache["statsd"] = []Address{
		{"127.0.0.1", 8125},
	}
	cache["mongo"] = []Address{
		{"127.0.0.1", 27017},
		{"192.168.10.123", 27017},
	}
}

func mustConnect() {
	if err := signal.WithExponentialBackoff(connect); err != nil {
		log.Printf("Giving up connecting %s", consulAddr)
		log.Fatal(err)
	}
}

func connect() error {
	config := api.DefaultConfig()
	config.Address = consulAddr
	c, err := api.NewClient(config)
	if err != nil {
		log.S("addr", consulAddr).Error(err)
		return err
	}
	fmt.Printf("connected to %s\n", consulAddr)
	consul = c
	if err := self(); err != nil {
		log.S("addr", consulAddr).Error(err)
		return err
	}
	// wait for dependencies to apear in consul
	if e, ok := os.LookupEnv(EnvWait); ok && e != "" {
		services := strings.Split(e, ",")
		for _, s := range services {
			if _, err := Services(s); err != nil {
				log.S("addr", consulAddr).S("service", s).Error(err)
				return err
			}
		}
	}
	return nil
}

func serviceName(fqdn, domain string) (string, string) {
	rx := regexp.MustCompile(fmt.Sprintf(`^(\S*)\.service\.*(\S*)*\.%s$`, domain))
	ms := rx.FindStringSubmatch(fqdn)
	if len(ms) < 2 {
		return fqdn, ""
	}
	if len(ms) > 2 {
		return ms[1], ms[2]
	}
	return ms[1], ""
}

func parseConsulServiceEntries(ses []*api.ServiceEntry) Addresses {
	srvs := []Address{}
	for _, se := range ses {
		addr := se.Service.Address
		if addr == "" {
			addr = se.Node.Address
		}
		srvs = append(srvs, Address{
			Address: addr,
			Port:    se.Service.Port,
		})
	}
	return srvs
}

func updateCache(name string, dc string, srvs []Address) {
	l.Lock()
	// log.Printf("updating cache for %s: %d records", name, len(srvs))
	cache[cacheKey(name, dc)] = srvs
	l.Unlock()
}

func invalidateCache(name string, dc string) {
	l.Lock()
	delete(cache, cacheKey(name, dc))
	l.Unlock()
}

func cacheKey(name string, dc string) string {
	if dc == "" {
		return name
	}
	return fmt.Sprintf("%s-%s", name, dc)
}

func monitor(name string, dc string, startIndex uint64) {
	wi := startIndex
	tries := 0
	for {
		qo := &api.QueryOptions{
			WaitIndex:         wi,
			WaitTime:          time.Minute * waitTimeMinutes,
			AllowStale:        true,
			RequireConsistent: false,
			Datacenter:        dc,
		}
		//log.Printf("querying Consul for %s with wait index: %d", name, wi)

		ses, qm, err := service(name, "", qo)
		if err != nil {
			tries++
			if tries == queryRetries {
				invalidateCache(name, dc)
				return
			}
			time.Sleep(time.Second * queryTimeoutSeconds)
			continue
		}
		tries = 0
		wi = qm.LastIndex
		updateCache(name, dc, parseConsulServiceEntries(ses))
	}
}

func service(service, tag string, qo *api.QueryOptions) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	ses, qm, err := consul.Health().Service(service, tag, false, qo)
	if err != nil {
		return nil, nil, err
	}
	// izbacujem servise koji imaju check koji nije ni "passing" ni "warning"
	var filteredSes []*api.ServiceEntry
loop:
	for _, se := range ses {
		for _, c := range se.Checks {
			if c.Status != "passing" && c.Status != "warning" {
				continue loop
			}
		}
		filteredSes = append(filteredSes, se)
	}
	return filteredSes, qm, nil
}

func query(name string, dc string) (Addresses, error) {
	//log.Printf("querying Consul for %s", name)
	qo := &api.QueryOptions{Datacenter: dc}
	ses, qm, err := service(name, "", qo)
	if err != nil {
		return nil, err
	}
	srvs := parseConsulServiceEntries(ses)
	if len(srvs) == 0 {
		return nil, fmt.Errorf(fmt.Sprintf("service %s not found in consul %s", name, consulAddr))
	}
	updateCache(name, dc, srvs)
	go func() {
		monitor(name, dc, qm.LastIndex)
	}()
	return srvs, nil
}

func srv(name string, dc string) (Addresses, error) {
	l.RLock()
	srvs, ok := cache[cacheKey(name, dc)]
	l.RUnlock()
	if ok && len(srvs) > 0 {
		// log.Printf("cache hit for %s: %d records", name, len(srvs))
		return srvs, nil
	}
	// log.Printf("cache miss for %s %v", name, srvs)
	srvs, err := query(name, dc)
	if err != nil {
		return nil, err
	}
	return srvs, nil
}

// Services retruns all services register in Consul.
func Services(name string) (Addresses, error) {
	sn, dc := serviceName(name, domain)
	return srv(sn, dc)
}

// Service will find one service in Consul cluster.
// Will randomly choose one if there are multiple register in Consul.
func Service(name string) (Address, error) {
	srvs, err := Services(name)
	if err != nil {
		return Address{}, err
	}
	srv := srvs[rand.Intn(len(srvs))]
	return srv, nil
}

// AgentService finds service on this (local) agent.
func AgentService(name string) (Address, error) {
	svcs, err := consul.Agent().Services()
	if err != nil {
		return Address{}, err
	}
	for _, svc := range svcs {
		//fmt.Printf("\t %#v\n", svc)
		if svc.Service == name {
			addr := svc.Address
			if addr == "" {
				addr = consulAddr
			}
			return Address{Address: addr, Port: svc.Port}, nil
		}
	}
	return Address{}, fmt.Errorf("service not found")
}

// Inspect Consul for configuration parameters.
func self() error {
	s, err := consul.Agent().Self()
	if err != nil {
		return err
	}
	c := s["Config"]
	domain = c["Domain"].(string)
	dc = c["Datacenter"].(string)
	nodeName = c["NodeName"].(string)
	advertiseAddr = c["AdvertiseAddr"].(string)
	bindAddr = c["BindAddr"].(string)
	return nil
}

// Call consul LockKey api function.
func LockKey(key string) (*api.Lock, error) {
	return consul.LockKey(key)
}

// NodeName returns Node name as defined in Consul.
func NodeName() string {
	return nodeName
}

// Dc returns datacenter name.
func Dc() string {
	return dc
}

// KV reads key from Consul key value storage.
func KV(key string) ([]byte, error) {
	kv := consul.KV()
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return nil, err
	}
	if pair == nil {
		return nil, fmt.Errorf("key not found")
	}
	return pair.Value, nil
}

// URL discovers host from url.
// If there are multiple services will randomly choose one.
func URL(url string) string {
	scheme, host, _, path, query := unpackURL(url)
	// log.S("url", url).S("host", host).Debug(fmt.Sprintf("should discover: %v", shouldDiscoverHost(host)))
	if !shouldDiscoverHost(host) {
		return url
	}
	srvs, err := Services(host)
	if err != nil {
		log.Error(err)
		return url
	}
	// log.I("len_srvs", len(srvs)).Debug("service entries")
	if len(srvs) == 0 {
		return url
	}
	srv := srvs[rand.Intn(len(srvs))]
	return packURL(scheme, srv.String(), "", path, query)
}

// shouldDiscoverHost - ima li smisla pitati consul za service discovery
func shouldDiscoverHost(name string) bool {
	parts := strings.Split(name, ".")
	if len(parts) == 1 {
		if parts[0] == "localhost" {
			return false
		}
		return true
	}
	return parts[len(parts)-1] == domain
}

func unpackURL(s string) (scheme, host, port, path string, query url.Values) {
	if strings.Contains(s, "//") {
		u, err := url.Parse(s)
		if err != nil {
			return
		}
		scheme = u.Scheme
		path = u.Path
		host = u.Host
		query = u.Query()
		h, p, err := net.SplitHostPort(u.Host)
		if err == nil {
			host = h
			port = p
		}
		return
	}

	host = s
	h, p, err := net.SplitHostPort(s)
	if err == nil {
		host = h
		port = p
	}
	return
}

func packURL(scheme, host, port, path string, query url.Values) (url string) {
	if scheme != "" {
		url = scheme + "://"
	}
	url += host
	if port != "" {
		url += ":" + port
	}
	url += path
	if len(query) > 0 {
		url += "?" + query.Encode()
	}

	return url
}

// MongoConnStr finds service mongo in consul and returns it in mongo connection string format.
func MongoConnStr() (string, error) {
	addrs, err := Services("mongo")
	if err != nil {
		return "", err
	}
	return strings.Join(addrs.String(), ","), nil
}
