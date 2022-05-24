package geo

import (
	"archive/tar"
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
	geoIp2 "github.com/oschwald/geoip2-golang"
)

const (
	GetFileRetryIntervalSeconds = 30
	GeoIpUrlLegacy           	= "http://updates.maxmind.com/app/update?license_key=LICENSE_KEY"
	GeoIp2Url					= "https://download.maxmind.com/app/geoip_download?edition_id=GeoIP2-Country&license_key=LICENSE_KEY&suffix=tar.gz"
	GeoIPCheckInterval          = 12 * 60 * time.Minute
)

const (
	GeoIpVersionLegacy	uint8 = 1
	GeoIpVersion2		uint8 = 2
)

var geoIPCheck *IPCheck
var lock sync.RWMutex

var whitelist = []string{"212.92.192.0/19"}

var versionUrl = map[uint8]string{
	GeoIpVersionLegacy: GeoIpUrlLegacy,
	GeoIpVersion2: 		GeoIp2Url,
}

// IPCheck is structure used for handling check of ip addresses
type IPCheck struct {
	file					string
	handleV1			*libgeo.GeoIP
	handleV2			*geoIp2.Reader
	cache 				map[string]bool
	IPWhitelist			[]*net.IPNet
	allowedCountryCodes []string
	sync.RWMutex
}

// Init initializes geoIPCheck
func Init(file string, version uint8, allowedCountryCodes []string) {
	lock.Lock()
	defer lock.Unlock()
	c, err := NewIPCheck(file, whitelist, version, allowedCountryCodes)
	if err != nil {
		log.Errorf("could not initialize Geo IP check from %s", file)
		return
	}
	geoIPCheck = c
}

// Default initializes IPCheck with default file if it exists
func Default() error {
	for _, p := range strings.Split(os.Getenv("GOPATH"), ":") {
		pth := path.Join(p, "src/pkg/geo/testGeoIP.dat")
		log.Info("trying to load Geo IP from: %s", pth)
		if _, err := os.Stat(pth); err == nil {
			Init(pth, GeoIpVersionLegacy, []string{"HR"})
			log.Info("loaded default Geo IP: %s", pth)
			return nil
		}
	}
	return fmt.Errorf("default Geo IP file doesnt exist")
}

// NewIPCheck initializes new IPCheck with specified file and version
func NewIPCheck(file string, whitelist []string, version uint8, allowedCountryCodes []string) (*IPCheck, error) {
	ipc := &IPCheck{
		file: file,
		cache: make(map[string]bool),
		IPWhitelist: initIPWhitelist(whitelist),
		allowedCountryCodes: allowedCountryCodes,
	}

	errorMsg := "error NewIpCheck could not open file %s, error: %s, version: %d\n"

	switch version {
	case GeoIpVersionLegacy:
		handleV1, err := libgeo.Load(file)
		if err != nil {
			log.Errorf(errorMsg, file, err, version)
			return nil, err
		}
		ipc.handleV1 = handleV1
		return ipc, nil
	case GeoIpVersion2:
		handleV2, err := geoIp2.Open(file)
		if err != nil {
			fmt.Printf(errorMsg, file, err, version)
			pth, err := os.Getwd()
			if err != nil {
				fmt.Println(err)
			}
			fmt.Println(pth)
			return nil, err
		}
		ipc.handleV2 = handleV2
		return ipc, nil
	default:
		return nil, fmt.Errorf("error NewIpCheck unknown geo ip version %d\n", version)
	}
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

	var cc string

	switch {
	case g.handleV1 != nil:
		location := g.handleV1.GetLocationByIP(ip)
		if location != nil {
			cc = location.CountryCode
		}
	case g.handleV2 != nil:
		i := net.ParseIP(ip)
		c, err :=g.handleV2.Country(i)

		if err != nil {
			log.Printf("error geo.Check on IP %s failed with error: %s", ip, err)
			return false
		}

		cc = c.Country.IsoCode
	default:
		log.Printf("geo ip handle not set")
		return false
	}

	if cc != "" {
		val := contains(g.allowedCountryCodes, cc)
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

func untarGeoIp2File(savePath string, r io.Reader) (int64, error) {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return 0, err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	foundDb := false
	for {
		header, err := tr.Next()
		switch {
		case err == io.EOF && !foundDb:
			return 0, fmt.Errorf("Database file not found in archive.")
		case err == io.EOF:
			return 0, nil
		case err != nil:
			return 0, err
		case header == nil:
			continue
		}

		if header.Typeflag != tar.TypeReg || !strings.Contains(header.Name, ".mmdb") {
			continue
		}

		foundDb = true

		f, err := os.OpenFile(savePath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
		if err != nil {
			return 0, err
		}
		n, err := io.Copy(f, tr)
		if err != nil {
			return 0, err
		}
		f.Close()

		return n, nil
	}
}

func getGeoIPFile(url, savePath string, version uint8) error {
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
	if version == GeoIpVersion2 {
		wn, err := untarGeoIp2File(savePath, res.Body)
		if err != nil {
			return err
		}
		log.Info("written %d bytes to %s", wn, savePath)
		return nil
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
func Maintain(savePath string, version uint8, licenseKey string, allowedCountryCodes []string) {
	url := versionUrl[version]
	url = strings.Replace(url, "LICENSE_KEY", licenseKey, 1)
	interval := GeoIPCheckInterval
	mt := func() time.Time {
		for {
			mt, err := checkGeoIPFile(savePath)
			if err != nil {
				log.Error(err)
				if err := getGeoIPFile(url, savePath, version); err != nil {
					log.Errorf("error getting GeoIP file: %v", err)
					time.Sleep(GetFileRetryIntervalSeconds * time.Second)
				}
			} else {
				Init(savePath, version, allowedCountryCodes)
				if geoIPCheck == nil {
					if err := getGeoIPFile(url, savePath, version); err != nil {
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
		if err := getGeoIPFile(url, savePath, version); err != nil {
			log.Errorf("error getting GeoIP file: %v", err)
		} else {
			newMt, err := checkGeoIPFile(savePath)
			if err != nil {
				log.Errorf("error: %v", err)
			} else {
				log.Info("curent geo ip: %v, new geo ip: %v", mt, newMt)
				if newMt.After(mt) {
					log.Info("loading geo ip")
					Init(savePath, version, allowedCountryCodes)
					mt = *newMt
				}
			}
		}
	}

}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
