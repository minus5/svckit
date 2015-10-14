package util

import (
	"bytes"
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"os/signal"
	"path"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/nu7hatch/gouuid"
)

var (
	sanitizeRx   = regexp.MustCompile("[^a-zA-Z0-9-]*")
	diacriticMap = map[string]string{
		"č": "c",
		"ć": "c",
		"đ": "dj",
		"ž": "z",
		"š": "s",
		"Č": "C",
		"Ć": "C",
		"Đ": "Dj",
		"Ž": "Z",
		"Š": "S",
		"ö": "oe",
		"Ö": "Oe",
		"ä": "ae",
		"Ä": "Ae",
		"ü": "ue",
		"Ü": "Ue",
		"ß": "ss",
		"ñ": "n",
		"Ñ": "N",
		"ç": "c",
		"Ç": "c",
		"á": "a",
		"é": "e",
		"í": "i",
		"ó": "o",
		"ú": "u",
		"à": "a",
		"è": "e",
		"ì": "i",
		"ò": "o",
		"ù": "u",
		"ë": "e",
		"ï": "i",
		"â": "a",
		"ê": "e",
		"î": "i",
		"ô": "o",
		"û": "u",
		"ã": "a",
		"õ": "o",
		"Á": "A",
		"É": "E",
		"Í": "I",
		"Ó": "O",
		"Ú": "U",
		"À": "A",
		"È": "E",
		"Ì": "I",
		"Ò": "O",
		"Ù": "U",
		"Ë": "E",
		"Ï": "I",
		"Â": "A",
		"Ê": "E",
		"Î": "I",
		"Ô": "O",
		"Û": "U",
		"Ã": "A",
		"Õ": "O",
	}
	diacriticReplacer = initDiacriticReplacer()
)

func initDiacriticReplacer() *strings.Replacer {
	args := []string{}
	for key, val := range diacriticMap {
		args = append(args, key, val)
	}
	return strings.NewReplacer(args...)
}

func Uuid() string {
	u4, err := uuid.NewV4()
	if err != nil {
		return ""
	}
	return strings.ToUpper(u4.String())
}

func InitLogger() {
	logFlag := log.LstdFlags | log.Lmicroseconds | log.Lshortfile
	log.SetFlags(logFlag)
}

func InitLoggerNoFile() {
	logFlag := log.LstdFlags | log.Lmicroseconds
	log.SetFlags(logFlag)
}

type Hash map[string]interface{}

func WaitForInterupt() {
	c := make(chan os.Signal, 1)
	//SIGINT je ctrl-C u shell-u, SIGTERM salje upstart kada se napravi sudo stop ...
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	<-c
}

func Hostname() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}
	return name
}

type StringArray []string

func (a *StringArray) Set(s string) error {
	*a = append(*a, s)
	return nil
}

func (a *StringArray) String() string {
	return strings.Join(*a, ",")
}

// return rounded version of x with prec precision.
func Round(x float64, prec int) float64 {
	var rounder float64
	pow := math.Pow(10, float64(prec))
	intermed := x * pow
	_, frac := math.Modf(intermed)
	intermed += .5
	x = .5
	if frac < 0.0 {
		x = -.5
		intermed -= 1
	}
	if frac >= x {
		rounder = math.Ceil(intermed)
	} else {
		rounder = math.Floor(intermed)
	}

	return rounder / pow
}

//UnixMilli - unix timestamp u milisekundama, pogodan za js
func UnixMilli() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//WriteFile - napravi direktorij (ako ne postoji) i sinimi tamo file
func WriteFile(file string, buf []byte) error {
	if err := makeDirFor(file); err != nil {
		return err
	}
	if err := ioutil.WriteFile(file, buf, 0644); err != nil {
		return err
	}
	return nil
}

func AppendToFile(file string, rdr io.Reader) error {
	if err := makeDirFor(file); err != nil {
		return err
	}
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, rdr)
	return err
}

func makeDirFor(file string) error {
	dir, _ := path.Split(file)
	return os.MkdirAll(dir, os.ModePerm)
}

func RetryDouble(first, max time.Duration, f func() bool) {
	Retry(first, max, 2, f)
}

func Retry(first, max time.Duration, base float64, f func() bool) {
	total := time.Duration(0)
	d := first
	i := 0
	for total < max {
		if !f() {
			nd := math.Pow(base, float64(i))
			d = time.Duration(float64(first.Nanoseconds())*nd) * time.Nanosecond
			log.Printf("sleeping %v", d)
			time.Sleep(d)
			total += d
			i++
		} else {
			return
		}
	}
}

func XMLPretty(data []byte) ([]byte, error) {
	b := &bytes.Buffer{}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	encoder := xml.NewEncoder(b)
	encoder.Indent("", "  ")
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			encoder.Flush()
			return b.Bytes(), nil
		}
		if err != nil {
			return nil, err
		}
		err = encoder.EncodeToken(token)
		if err != nil {
			return nil, err
		}
	}
}

func diacriticReplace(s string) string {
	for key, val := range diacriticMap {
		s = strings.Replace(s, key, val, -1)
	}
	return s
}

func Sanitize(s string) string {
	s = diacriticReplacer.Replace(s)
	return sanitizeRx.ReplaceAllString(s, "")
}
