package geo

import (
	"compress/gzip"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/nranchev/go-libGeoIP"
)

const (
	GET_FILE_RETRY_INTERVAL_SECONDS = 30
)

var geoIpCheck *IpCheck
var lock sync.RWMutex

type IpCheck struct {
	file       string
	handle     *libgeo.GeoIP
	cache      map[string]bool
	cacheMutex sync.Mutex
}

func Init(file string) {
	lock.Lock()
	defer lock.Unlock()
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
	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()
	res, err := http.Get(url)
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

func checkGeoIpFile(path string) (time.Time, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("geo ip file doesn't exist")
		} else {
			log.Printf("geo ip file unreadable")
		}
	}
	return fi.ModTime(), nil
}

func Maintain(url, savePath string, interval time.Duration) {

	mt := func() time.Time {
		for {
			mt, err := checkGeoIpFile(savePath)
			if err != nil {
				if err := getGeoIpFile(url, savePath); err != nil {
					log.Printf("error getting GeoIP file: %v", err)
					time.Sleep(GET_FILE_RETRY_INTERVAL_SECONDS * time.Second)
				}
			} else {
				Init(savePath)
				return mt
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
				if newMt.After(mt) {
					Init(savePath)
					mt = newMt
				}
			}
		}
	}

}
