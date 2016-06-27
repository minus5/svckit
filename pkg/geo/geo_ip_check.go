package geo

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"pkg/httpu"
	"strings"
	"sync"
	"time"

	"github.com/nranchev/go-libGeoIP"
)

const (
	GetFileRetryIntervalSeconds = 30
	GeoIpUrl                    = "http://updates.maxmind.com/app/update?license_key=MHJmPlgC696x"
	GeoIpCheckInterval          = 12 * 60 * time.Minute
)

var geoIpCheck *IpCheck
var lock sync.RWMutex

type IpCheck struct {
	file       string
	handle     *libgeo.GeoIP
	cache      map[string]bool
	cacheMutex sync.RWMutex
}

func Init(file string) {
	lock.Lock()
	defer lock.Unlock()
	c, err := NewIpCheck(file)
	if err != nil {
		log.Printf("could not initialize Geo IP check from %s", file)
		return
	}
	geoIpCheck = c
}

func Default() error {
	for _, p := range strings.Split(os.Getenv("GOPATH"), ":") {
		path := path.Join(p, "src/pkg/geo/testGeoIP.dat")
		log.Printf("trying to load Geo IP from: %s", path)
		if _, err := os.Stat(path); err == nil {
			Init(path)
			log.Printf("loaded default Geo IP: %s", path)
			return nil
		}
	}
	return fmt.Errorf("default Geo IP file doesnt exist")
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
	g.cacheMutex.RLock()
	val, ok := g.cache[ip]
	g.cacheMutex.RUnlock()
	if ok {
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
	lock.RLock()
	defer lock.RUnlock()
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

func getGeoIpFile(url, savePath string) error {
	dir := filepath.Dir(savePath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	res, err := httpu.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	lastModified, err := time.Parse(time.RFC1123, res.Header.Get("Last-Modified"))
	if err != nil {
		log.Printf("unable to parse modification time of GeoIP database")
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
	log.Printf("written %d bytes to %s", n, savePath)
	out.Close()
	log.Print(lastModified)
	if err := os.Chtimes(savePath, time.Now(), lastModified); err != nil {
		return err
	}
	return nil
}

func checkGeoIpFile(path string) (*time.Time, error) {
	log.Printf("checking geo ip at %s", path)
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("geo ip file doesn't exist")
		} else {
			return nil, fmt.Errorf("geo ip file unreadable")
		}
	}
	log.Printf("found geo ip at %s", path)
	tm := new(time.Time)
	*tm = fi.ModTime()
	return tm, nil
}

func Maintain(savePath string) {
	url := GeoIpUrl
	interval := GeoIpCheckInterval
	mt := func() time.Time {
		for {
			mt, err := checkGeoIpFile(savePath)
			if err != nil {
				log.Println(err)
				if err := getGeoIpFile(url, savePath); err != nil {
					log.Printf("error getting GeoIP file: %v", err)
					time.Sleep(GetFileRetryIntervalSeconds * time.Second)
				}
			} else {
				Init(savePath)
				if geoIpCheck == nil {
					if err := getGeoIpFile(url, savePath); err != nil {
						log.Printf("error getting GeoIP file: %v", err)
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
		if err := getGeoIpFile(url, savePath); err != nil {
			log.Printf("error getting GeoIP file: %v", err)
		} else {
			newMt, err := checkGeoIpFile(savePath)
			if err != nil {
				log.Printf("error: %v", err)
			} else {
				log.Printf("curent geo ip: %v, new geo ip: %v", mt, newMt)
				if newMt.After(mt) {
					log.Printf("loading geo ip")
					Init(savePath)
					mt = *newMt
				}
			}
		}
	}

}
