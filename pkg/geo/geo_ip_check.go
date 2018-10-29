package geo

import (
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/minus5/svckit/log"
	"github.com/nranchev/go-libGeoIP"
)

const (
	GetFileRetryIntervalSeconds = 30
	GeoIPURL                    = "http://updates.maxmind.com/app/update?license_key=MHJmPlgC696x"
	GeoIPCheckInterval          = 12 * 60 * time.Minute
)

var geoIPCheck *IPCheck
var lock sync.RWMutex

var whitelist = []string{"212.92.196.34/32"}

// IPCheck is structure used for handling check of ip addresses
type IPCheck struct {
	file        string
	handle      *libgeo.GeoIP
	cache       map[string]bool
	IPWhitelist []*net.IPNet
	sync.RWMutex
}

// Init initializes geoIPCheck
func Init(file string) {
	lock.Lock()
	defer lock.Unlock()
	c, err := NewIPCheck(file, whitelist)
	if err != nil {
		log.Errorf("could not initialize Geo IP check from %s", file)
		return
	}
	geoIPCheck = c
}

// Default initializes IPCheck with default file if it exists
func Default() error {
	for _, p := range strings.Split(os.Getenv("GOPATH"), ":") {
		path := path.Join(p, "src/pkg/geo/testGeoIP.dat")
		log.Info("trying to load Geo IP from: %s", path)
		if _, err := os.Stat(path); err == nil {
			Init(path)
			log.Info("loaded default Geo IP: %s", path)
			return nil
		}
	}
	return fmt.Errorf("default Geo IP file doesnt exist")
}

// NewIPCheck initializes new IPCheck with specified file
func NewIPCheck(file string, whitelist []string) (*IPCheck, error) {
	handle, err := libgeo.Load(file)
	if err != nil {
		log.Errorf("error NewIpCheck could not open file %s, error: %s", file, err)
		return nil, err
	}

	return &IPCheck{file: file, handle: handle, cache: make(map[string]bool), IPWhitelist: initIPWhitelist(whitelist)}, nil
}

// Check checks if IP is from specified area
func (g *IPCheck) Check(ip string) (ret bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("error geo.Check failed: %v", r)
			ret = true
		}
	}()
	if isLocalAddress(ip) {
		return true
	}
	g.RLock()
	val, ok := g.cache[ip]
	g.RUnlock()
	if ok {
		return val
	}

	if g.checkIPWhitelist(ip) {
		g.Lock()
		defer g.Unlock()
		g.cache[ip] = true
		return true
	}

	loc := g.handle.GetLocationByIP(ip)
	if loc != nil {
		val := loc.CountryCode == "HR"
		g.Lock()
		defer g.Unlock()
		g.cache[ip] = val
		return val
	}
	log.Printf("warn ip %s not found in database", ip)
	return false
}

func (g *IPCheck) checkIPWhitelist(ip string) bool {
	netip := net.ParseIP(ip)
	for _, wl := range g.IPWhitelist {
		if wl.Contains(netip) {
			return true
		}
	}
	return false
}

// IpOk does check of ip if IPCheck is initialized
func IpOk(ip string) bool {
	lock.RLock()
	defer lock.RUnlock()
	if geoIPCheck == nil {
		log.Printf("WARN geo ip check not activated")
		return true
	}
	return geoIPCheck.Check(ip)
}

func isLocalAddress(ip string) bool {
	i := net.ParseIP(ip)
	if i != nil {
		ip4 := i.To4()
		if (ip4[0] == 192 && ip4[1] == 168) ||
			(ip4[0] == 10) ||
			(ip4[0] == 127 && ip4[1] == 0 && ip4[2] == 0) ||
			(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) {
			return true
		}
	}
	return false
}

func getGeoIPFile(url, savePath string) error {
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	lastModified, err := time.Parse(time.RFC1123, res.Header.Get("Last-Modified"))
	if err != nil {
		log.Errorf("unable to parse modification time of GeoIP database")
		lastModified = time.Now()
	}
	gzReader, err := gzip.NewReader(res.Body)
	if err != nil {
		return err
	}
	defer gzReader.Close()
	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()
	n, err := io.Copy(out, gzReader)
	if err != nil {
		return err
	}
	log.Info("written %d bytes to %s", n, savePath)
	out.Close()
	if err := os.Chtimes(savePath, time.Now(), lastModified); err != nil {
		return err
	}
	return nil
}

func checkGeoIPFile(path string) (*time.Time, error) {
	log.Printf("checking geo ip at %s", path)
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("geo ip file doesn't exist")
		}
		return nil, fmt.Errorf("geo ip file unreadable")
	}

	log.Info("found geo ip at %s", path)
	tm := new(time.Time)
	*tm = fi.ModTime()
	return tm, nil
}

func initIPWhitelist(whitelist []string) []*net.IPNet {
	wl := make([]*net.IPNet, 0)
	for _, ip := range whitelist {
		_, net, err := net.ParseCIDR(ip)
		if err != nil {
			log.Error(err)
			continue
		}
		wl = append(wl, net)
	}
	return wl

}

// Maintain keeps track of ip database and downloads new version when available
func Maintain(savePath string) {
	url := GeoIPURL
	interval := GeoIPCheckInterval
	mt := func() time.Time {
		for {
			mt, err := checkGeoIPFile(savePath)
			if err != nil {
				log.Error(err)
				if err := getGeoIPFile(url, savePath); err != nil {
					log.Errorf("error getting GeoIP file: %v", err)
					time.Sleep(GetFileRetryIntervalSeconds * time.Second)
				}
			} else {
				Init(savePath)
				if geoIPCheck == nil {
					if err := getGeoIPFile(url, savePath); err != nil {
						log.Errorf("error getting GeoIP file: %v", err)
						time.Sleep(GetFileRetryIntervalSeconds * time.Second)
					}
					continue
				}
				return *mt
			}
		}
	}()

	for {
		time.Sleep(interval)
		if err := getGeoIPFile(url, savePath); err != nil {
			log.Errorf("error getting GeoIP file: %v", err)
		} else {
			newMt, err := checkGeoIPFile(savePath)
			if err != nil {
				log.Errorf("error: %v", err)
			} else {
				log.Info("curent geo ip: %v, new geo ip: %v", mt, newMt)
				if newMt.After(mt) {
					log.Info("loading geo ip")
					Init(savePath)
					mt = *newMt
				}
			}
		}
	}

}
