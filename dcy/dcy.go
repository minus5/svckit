package dcy

import (
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mnu5/svckit/env"
	"github.com/mnu5/svckit/log"
	"github.com/mnu5/svckit/signal"

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
	consul      *api.Client
	l           sync.RWMutex
	cache       = map[string]Addresses{}
	subscribers = map[string][]func(Addresses){}

	domain     string
	dc         string
	nodeName   string
	consulAddr = localConsulAdr
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

func (a Address) Equal(a2 Address) bool {
	return a.Address == a2.Address && a.Port == a2.Port
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

func (a Addresses) Equal(a2 Addresses) bool {
	if len(a) != len(a2) {
		return false
	}
	for _, d := range a {
		found := false
		for _, d2 := range a2 {
			if d.Equal(d2) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (a Addresses) Contains(a2 Address) bool {
	for _, a1 := range a {
		if a1.Equal(a2) {
			return true
		}
	}
	return false
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
	if consulAddr == "--" {
		return
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
	domain = "sd"
	dc = "dev"
	nodeName = "node01"
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
	cache["nsqlookupd-http"] = []Address{
		{"127.0.0.1", 4161},
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

func ConnectTo(addr string) error {
	if consul != nil {
		return nil
	}
	consulAddr = addr
	return signal.WithExponentialBackoff(connect)
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

func updateCache(tag, name, dc string, srvs Addresses) {
	l.Lock()
	defer l.Unlock()
	key := cacheKey(tag, name, dc)
	if srvs2, ok := cache[key]; ok {
		if srvs2.Equal(srvs) {
			return
		}
	}
	cache[key] = srvs
	notify(name, srvs)

}

func invalidateCache(tag, name, dc string) {
	l.Lock()
	defer l.Unlock()
	delete(cache, cacheKey(tag, name, dc))
}

func cacheKey(tag, name, dc string) string {
	var key string
	if tag != "" {
		key = fmt.Sprintf("%s-", tag)
	}
	if dc == "" {
		return fmt.Sprintf("%s%s", key, name)
	}
	return fmt.Sprintf("%s%s-%s", key, name, dc)
}

func monitor(tag, name, dc string, startIndex uint64) {
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
		ses, qm, err := service(name, tag, qo)
		if err != nil {
			tries++
			if tries == queryRetries {
				invalidateCache(tag, name, dc)
				return
			}
			time.Sleep(time.Second * queryTimeoutSeconds)
			continue
		}
		tries = 0
		wi = qm.LastIndex
		updateCache(tag, name, dc, parseConsulServiceEntries(ses))
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

func query(tag, name, dc string) (Addresses, error) {
	qo := &api.QueryOptions{Datacenter: dc}
	ses, qm, err := service(name, tag, qo)
	if err != nil {
		return nil, err
	}
	srvs := parseConsulServiceEntries(ses)
	if len(srvs) == 0 {
		return nil, fmt.Errorf("service %s not found in consul %s", name, consulAddr)
	}
	updateCache(tag, name, dc, srvs)
	go func() {
		monitor(tag, name, dc, qm.LastIndex)
	}()
	return srvs, nil
}

func srvQuery(tag, name string, dc string) (Addresses, error) {
	l.RLock()
	srvs, ok := cache[cacheKey(tag, name, dc)]
	l.RUnlock()
	if ok && len(srvs) > 0 {
		return srvs, nil
	}
	srvs, err := query(tag, name, dc)
	if err != nil {
		return nil, err
	}
	return srvs, nil
}

func srv(tag, name string, dc string) (Addresses, error) {
	srvs, err := srvQuery(tag, name, dc)
	if err == nil {
		return srvs, nil
	}

	nameNomad := strings.Replace(name, "_", "-", -1)
	srvs, err = srvQuery(tag, nameNomad, dc)
	if err != nil {
		return nil, err
	}
	return srvs, nil
}

// Services retruns all services register in Consul.
func Services(name string) (Addresses, error) {
	sn, dc := serviceName(name, domain)
	return srv("", sn, dc)
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

// ServiceInDc will find one service in Consul claster for specified datacenter
func ServiceInDc(name, dc string) (Address, error) {
	return ServiceInDcByTag("", name, dc)
}

// ServiceInDcByTag will find one service in consul claster with tag for specified datacenter
func ServiceInDcByTag(tag, name, dc string) (Address, error) {
	srvs, err := srv(tag, name, dc)
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
	cfg := s["Config"]
	version := cfg["Version"].(string)
	dc = cfg["Datacenter"].(string)
	nodeName = cfg["NodeName"].(string)
	if strings.HasPrefix(version, "0.") {
		domain = cfg["Domain"].(string)
	} else {
		if dcfg := s["DebugConfig"]; dcfg != nil {
			domain = dcfg["DNSDomain"].(string)
		}
	}
	return nil
}

// Call consul LockKey api function.
func LockKey(key string) (*api.Lock, error) {
	opts := &api.LockOptions{
		Key:          key,
		LockWaitTime: 5 * time.Second,
	}
	return consul.LockOpts(opts)
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
func KV(key string) (string, error) {
	kv := consul.KV()
	pair, _, err := kv.Get(key, nil)
	if err != nil {
		return "", err
	}
	if pair == nil {
		return "", fmt.Errorf("key not found")
	}
	return string(pair.Value), nil
}

// KVs read keys from Consul key value storage.
func KVs(key string) (map[string]string, error) {
	kv := consul.KV()
	entries, _, err := kv.List(key, nil)
	if err != nil {
		return nil, err
	}
	if entries == nil {
		return nil, fmt.Errorf("key not found")
	}
	m := make(map[string]string)
	for _, e := range entries {
		k := strings.TrimPrefix(e.Key, key)
		k = strings.TrimPrefix(k, "/")
		m[k] = string(e.Value)
	}
	return m, nil
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
func MongoConnStr(names ...string) (string, error) {
	name := "mongo"
	if len(names) > 0 {
		name = names[0]
	}
	addrs, err := Services(name)
	if err != nil {
		return "", err
	}
	return strings.Join(addrs.String(), ","), nil
}

// Agent returns ref to consul agent.
// Only for use in sr package below.
func Agent() *api.Agent {
	return consul.Agent()
}

// MustConnect connects to real consul.
// Useful in tests, when dcy is started in test mode to force to connect to real consul.
func MustConnect() {
	mustConnect()
}

// Subscribe on service changes.
// Changes in Consul for service `name` will be passed to handler.
func Subscribe(name string, handler func(Addresses)) {
	_, err := Service(name) // query for service so monitor goroutine starts
	if err != nil {
		log.Error(err)
	}
	l.Lock()
	defer l.Unlock()
	a := subscribers[name]
	if a == nil {
		a = make([]func(Addresses), 0)
	}
	a = append(a, handler)
	subscribers[name] = a
}

func notify(name string, srvs Addresses) {
	if s, ok := subscribers[name]; ok {
		for _, h := range s {
			h(srvs)
		}
	}
}

// Unsubscribe from service changes.
func Unsubscribe(name string, handler func(Addresses)) {
	l.Lock()
	defer l.Unlock()
	a := subscribers[name]
	if a == nil {
		return
	}
	for i, h := range a {
		sf1 := reflect.ValueOf(h)
		sf2 := reflect.ValueOf(handler)
		if sf1.Pointer() == sf2.Pointer() {
			a = append(a[:i], a[i+1:]...)
			break
		}
	}
	subscribers[name] = a
}
