//Nadogradnja na postojeci go logger.
//Kada ukljucim ovaj package on se na init-u zakaci na standarni logger
//i outpu mu zmajeni svojim. Ako ne konfiguriram nsq nista se ne dogadja
//radi kao standard logger.
//Ako konfiguriram nsq onda parsa logove, napravi json i posalje na nsq.
//Jedan od nacina konfiguriranja je da postavim enviroment varijable.
//Ideja je da imam sto manje promjena u aplikaciji, a da izvucem logove na nsq.
package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"pkg/util/nsqu"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/bitly/go-nsq"
)

const (
	//imena enviroment varijabli koje se koriste
	envNsqdAddress = "log_nsqd_address"
	envDc          = "dc"
	envNode        = "node"

	logRegexpDate = `(?P<date>[0-9]{4}/[0-9]{2}/[0-9]{2})?[ ]?`
	logRegexpTime = `(?P<time>[0-9]{2}:[0-9]{2}:[0-9]{2}(\.[0-9]+)?)?[ ]?`
	logRegexpFile = `(?P<file>.+?:[0-9]+)?`
	logRegexpMsg  = `(: )?(\[(?P<level>.*)\]+)?(?P<msg>(?s:.*))\n?`
)

var (
	logRegexp     = regexp.MustCompile(logRegexpDate + logRegexpTime + logRegexpFile + logRegexpMsg)
	stdOutEnabled = true
	nsqdEnabled   = false
	nsqd          *nsq.Producer
	nsqTopic      = "logs"
	dc            = os.Getenv("dc")
	node          = os.Getenv("node")
	host          string
	app           = path.Base(os.Args[0])
)

/*LogPanic - panic po default otidje na stderr.
Ako ga zelim uhvatiti i logirati panic onda u main() stavim:
   defer func() {
    if r := recover(); r != nil {
 	    logger.LogPanic(r)
    }
	}()
*/
func LogPanic(r interface{}) {
	log.Printf("[PANIC] %v %s", r, debug.Stack())
	os.Exit(-1)
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	log.SetOutput(&stdLoggerOutput{})
	host, _ = os.Hostname()
	if v, ok := os.LookupEnv(envNsqdAddress); ok {
		InitNsq(v, "")
	}
}

//DisableStdout - zabrani pisanje logova na stdout
func DisableStdout() {
	stdOutEnabled = false
}

//InitNsq - inicijalizacija nsq logiranje. Ako se ne pozove nece se logirati na nsq.
func InitNsq(nsqdAddress, topic string) error {
	if topic != "" {
		nsqTopic = topic
	}
	var err error
	nsqd, err = nsqu.NewConnection(nil, nsqdAddress).NewProducer()
	if err != nil {
		return err
	}
	nsqdEnabled = true
	return nil
}

func subexps(line []byte) map[string]string {
	m := logRegexp.FindSubmatch(line)
	if len(m) < len(logRegexp.SubexpNames()) {
		return map[string]string{}
	}
	result := map[string]string{}
	for i, name := range logRegexp.SubexpNames() {
		if name != "" {
			result[name] = string(m[i])
		}
	}
	return result
}

type stdLoggerOutput struct {
	no int
}

func (o *stdLoggerOutput) Write(p []byte) (int, error) {
	o.no++
	if stdOutEnabled {
		fmt.Print(string(p))
	}
	if nsqdEnabled {
		m := toMessage(p, o.no)
		buf, err := json.Marshal(m)
		if err != nil {
			fmt.Printf("parseMessage error: %s", err)
		}
		if err := nsqd.Publish(nsqTopic, buf); err != nil {
			fmt.Printf("nsqd.Publish error: %s", err)
		}
	}
	return len(p), nil
}

func toMessage(p []byte, no int) message {
	parts := subexps(p)
	if l := parts["level"]; l != "" {
		parts["level"] = strings.ToLower(l)
	}

	d := message{
		//PAZI da ne moram parsati vrijeme odlucio sam sa da uzmem trenutno
		Time:    time.Now(),
		File:    parts["file"],
		Level:   strings.ToLower(parts["level"]),
		Message: strings.TrimSpace(parts["msg"]),
		No:      no,
		Dc:      dc,
		Node:    node,
		Host:    host,
		App:     app,
	}
	if d.Level == "" {
		d.Level = "debug"
		if strings.Contains(d.Message, "error") {
			d.Level = "error"
		}
	}

	return d
}

type message struct {
	Dc      string    `json:"dc,omitempty"`
	Node    string    `json:"node,omitempty"`
	Host    string    `json:"host,omitempty"`
	App     string    `json:"app,omitempty"`
	Time    time.Time `json:"time,omitempty"`
	File    string    `json:"file,omitempty"`
	Level   string    `json:"level,omitempty"`
	Message string    `json:"msg,omitempty"`
	No      int       `json:"no,omitempty"`
}
