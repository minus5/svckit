package log

import (
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"os"

	"github.com/mnu5/svckit/env"

	golog "log"
)

const (
	SyslogServiceName = "syslog"
	EnvSyslog         = "SVCKIT_LOG_SYSLOG"
	EnvDisableDebug   = "SVCKIT_LOG_DISABLE_DEBUG"
	EnvNode           = "node"
)

var (
	out                  io.Writer
	prefix               []byte
	debugLogLevelEnabled = true
)

type stdLibOutput struct{}

func (o *stdLibOutput) Write(p []byte) (int, error) {
	if len(p) > 0 {
		//izbaci zadnji znak (\n)
		p = p[0 : len(p)-1]
	}
	msg := string(p)
	level, msg := splitLevelMessage(msg)
	if level == LevelDebug && !debugLogLevelEnabled {
		return len(p), nil
	}
	a := newAgregator(5)
	a.level = level
	a.msg = msg
	a.write()
	return len(p), nil
}

/* TODO
- opcija da iskljuci logiranje file-a
- preusmjeri na syslog
- pazi na keys koji su vec zauzeti kada netko pokusava u njih upisati nesto, dodaj im neki prefix (reserved keys)
- sync.Pool iskoristi da ima vise buffera pa da ne mora nikada raditi lock
*/
func init() {
	out = os.Stderr

	hostname := env.Hostname()
	if node := os.Getenv(EnvNode); node != "" {
		hostname = node
	}

	//prefix za sve logove
	p := fmt.Sprintf(`"host":"%s", "app":"%s"`, hostname, env.AppName())
	prefix = []byte(p)

	// preusmjeri go standard lib logger kroz mene
	golog.SetFlags(0)
	golog.SetOutput(&stdLibOutput{})
	initSyslog()
	initLogLevel()
}

func initSyslog() {
	env, ok := os.LookupEnv(EnvSyslog)
	if !ok || (env == "0") || (env == "false") {
		return
	}
	if len(env) > 5 {
		setSyslogOutput(env)
		return
	}

	setSyslogOutput("127.0.0.1:514")
}

func initLogLevel() {
	env, ok := os.LookupEnv(EnvDisableDebug)
	if !ok || (env == "0") || (env == "false") || (env == "") {
		return
	}
	DisableDebug()
}

// DisableDebug do not log Debug messages
func DisableDebug() {
	debugLogLevelEnabled = false
}

func setSyslogOutput(addr string) {
	sys, err := syslog.Dial("udp", addr, syslog.LOG_LOCAL5, env.AppName())
	if err != nil {
		//For udp err is not raised if server don't exists.
		return
	}
	SetOutput(sys)
}

//SetOutput sets output for logs.
//Usefull for redirecting all logs to syslog server.
func SetOutput(o io.Writer) {
	out = o
}

func Discard() {
	SetOutput(ioutil.Discard)
	golog.SetOutput(ioutil.Discard)
}

func Printf(format string, v ...interface{}) {
	if !debugLogLevelEnabled {
		return
	}
	level, msg := splitLevelMessage(format)
	a := newAgregator(3)
	a.level = level
	a.msg = sprintf(msg, v...)
	a.write()
}

func Debug(msg string, v ...interface{}) {
	newAgregator(4).Debug(sprintf(msg, v...))
}

func Info(msg string, v ...interface{}) {
	newAgregator(4).Info(sprintf(msg, v...))
}

func Error(err error) {
	newAgregator(4).Error(err)
}

func Errorf(msg string, v ...interface{}) {
	newAgregator(4).Error(fmt.Errorf(msg, v...))
}

func Notice(msg string, v ...interface{}) {
	newAgregator(4).Notice(sprintf(msg, v...))
}

func Fatal(err error) {
	newAgregator(4).Fatal(err)
}

func Fatalf(msg string, v ...interface{}) {
	newAgregator(4).Fatal(fmt.Errorf(msg, v...))
	os.Exit(-1)
}

func sprintf(msg string, v ...interface{}) string {
	if len(v) != 0 {
		return fmt.Sprintf(msg, v...)
	}
	return msg
}

func I(key string, val int) *Agregator {
	return newAgregator(3).I(key, val)
}

func F(key string, val float64, prec int) *Agregator {
	return newAgregator(3).F(key, val, prec)
}

func S(key string, val string) *Agregator {
	return newAgregator(3).S(key, val)
}

func J(key string, val []byte) *Agregator {
	return newAgregator(3).J(key, val)
}

func B(key string, val bool) *Agregator {
	return newAgregator(3).B(key, val)
}

func Jc(key string, val []byte) *Agregator {
	return newAgregator(3).Jc(key, val)
}

func Write(buf []byte) {
	out.Write(buf)
}
