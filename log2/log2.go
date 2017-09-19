package log2

import (
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"os"

	"github.com/minus5/svckit/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	SyslogServiceName = "syslog"
	EnvSyslog         = "SVCKIT_LOG_SYSLOG"
	EnvDisableDebug   = "SVCKIT_LOG_DISABLE_DEBUG"
)

var (
	out io.Writer
	a   Agregator
	cfg zap.Config
)

type stdLibOutput struct{}

//Write returns size of the parameter and error
func (o *stdLibOutput) Write(p []byte) (int, error) {
	if len(p) > 0 {
		//izbaci zadnji znak (\n)
		p = p[0 : len(p)-1]
	}

	msg := string(p)
	level, msg := splitLevelMessage(msg)

	a := newAgregator(3)
	a.print(level, msg)

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

	initSyslog()

	cfg = zap.NewProductionConfig()
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.CallerKey = "file"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	//kako bi se implementiralo pisanje u syslog potrebno je promijeniti rad
	//nekih funkcija iz zap biblioteke (uvodim fromZap.go)
	a.zlog = build(cfg, out, zap.Fields(
		zap.String("host", env.Hostname()),
		zap.String("app", env.AppName()),
	))

	//pribaceno u interni build
	//a.zlog = logger
}

// initSyslog gets env variable
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

// setSyslogOutput sets syslog as output
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

// Discard discardes log output
func Discard() {
	SetOutput(ioutil.Discard)
}

// Printf addes msg to log and writes log
func Printf(format string, v ...interface{}) {
	level, msg := splitLevelMessage(format)
	a := newAgregator(1)
	a.print(level, msg)
}

// Debug function receives message, sets: "level":"debug" and "msg":"message"
// in log buffer and then prints it
func Debug(msg string, v ...interface{}) {
	newAgregator(2).Debug(sprintf(msg, v...))
}

// Info function receives message, sets: "level":"info" and "msg":"message"
// in log buffer and then prints it
func Info(msg string, v ...interface{}) {
	newAgregator(2).Info(sprintf(msg, v...))
}

// Error function receives error, sets: "level":"error" and "msg":"error"
// if error != nil or "msg":"" if error == nill
// in log buffer and then prints it
func Error(err error) {
	newAgregator(2).Error(err)
}

//Errorf function receives message, sets "level":"error" and "msg":"message"
//in log buffer and then prints it
func Errorf(msg string, v ...interface{}) {
	newAgregator(2).Error(fmt.Errorf(msg, v...))
}

// Notice function receives message, sets: "level":"notice" and "msg":"message"
// in log buffer and then prints it
func Notice(msg string, v ...interface{}) {
	newAgregator(2).Notice(sprintf(msg, v...))
}

// Fatal function receives error, sets: "level":"fatal" and "msg":"error"
// if error != nil or "msg":"" if error == nill
// in log buffer, prints buffer and then exits
func Fatal(err error) {
	newAgregator(2).Fatal(err)
}

// Fatalf function receives error, sets: "level":"fatal" and "msg":"error"
// in log buffer, prints buffer and then exits
func Fatalf(msg string, v ...interface{}) {
	newAgregator(2).Fatal(fmt.Errorf(msg, v...))
	os.Exit(-1)
}

//sprintf returns msg or everything
func sprintf(msg string, v ...interface{}) string {
	if len(v) != 0 {
		return fmt.Sprintf(msg, v...)
	}
	return msg
}

//B finds and handles key-value pairs with boolean value
func B(key string, val bool) *Agregator {
	return newAgregator(1).B(key, val)
}

//I finds and handles key-value pairs with int value
func I(key string, val int) *Agregator {
	return newAgregator(1).I(key, val)
}

//F finds and handles key-value pairs with float64 value
func F(key string, val float64, prec int) *Agregator {
	return newAgregator(1).F(key, val, prec)
}

//S finds and handles key-value pairs with string value
func S(key string, val string) *Agregator {
	return newAgregator(1).S(key, val)
}

//J finds and handles key-value pairs with []byte value
func J(key string, val []byte) *Agregator {
	return newAgregator(1).J(key, val)
}

// Jc finds and handles key-value pairs with []byte value with check
func Jc(key string, val []byte) *Agregator {
	return newAgregator(1).Jc(key, val)
}

// Write sends byte array into output
func Write(buf []byte) {
	out.Write(buf)
}
