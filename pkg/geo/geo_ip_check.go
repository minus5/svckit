package geo

import (
	"log"
	"net"
	"sync"

	"github.com/nranchev/go-libGeoIP"
)

var geoIpCheck *IpCheck

type IpCheck struct {
	file       string
	handle     *libgeo.GeoIP
	cache      map[string]bool
	cacheMutex sync.Mutex
}

func Init(file string) {
	geoIpCheck, _ = NewIpCheck(file)
}

func NewIpCheck(file string) (*IpCheck, error) {
	handle, err := libgeo.Load(file)
	if err != nil {
		log.Printf("error NewIpCheck could not open file %s, error: %s", file, err)
		return nil, err
	}
	return &IpCheck{file: file, handle: handle, cache: make(map[string]bool)}, nil
}

func (g *IpCheck) Check(ip string) (ret bool) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("error geo.Check failed: %v", r)
			ret = true
		}
	}()
	if isLocalAddress(ip) {
		return true
	}
	if val, ok := g.cache[ip]; ok {
		return val
	}
	loc := g.handle.GetLocationByIP(ip)
	if loc != nil {
		val := loc.CountryCode == "HR"
		g.cacheMutex.Lock()
		defer g.cacheMutex.Unlock()
		g.cache[ip] = val
		return val
	}
	log.Printf("warn ip %s not found in database", ip)
	return false
}

func IpOk(ip string) bool {
	if geoIpCheck == nil {
		log.Printf("WARN geo ip check not activated")
		return true
	}
	return geoIpCheck.Check(ip)
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
