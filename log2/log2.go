package log2

import (
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/minus5/svckit/env"
)

const (
	SyslogServiceName = "syslog"
	EnvSyslog         = "SVCKIT_LOG_SYSLOG"
	EnvDisableDebug   = "SVCKIT_LOG_DISABLE_DEBUG"
)

var (
	out                  io.Writer
	cfg                  zap.Config
	prefix               []byte
	debugLogLevelEnabled = true
)

type stdLibOutput struct{}

//Write returns size of the parameter and error
func (o *stdLibOutput) Write(p []byte) (int, error) {
	if !debugLogLevelEnabled {
		return len(p), nil
	}
	if len(p) > 0 {
		//izbaci zadnji znak (\n)
		p = p[0 : len(p)-1]
	}

	msg := string(p)
	level, msg := splitLevelMessage(msg)

	a := newAgregator(5)
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

	cfg = zap.NewDevelopmentConfig()
	cfg.EncoderConfig.TimeKey = "time"
	cfg.EncoderConfig.CallerKey = "file"
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.MessageKey = "msg"
	cfg.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.Development = true
	cfg.Encoding = "json"

	//initSyslog()
	//initLogLevel()
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

//pozva disabledebug
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

// uspostavlja vezu sa serveron
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
// kako maknit sadrzaj ouputa za zap posto nemogu pristupit direktno
// zap-ovon outputpath-u. Tip podatka je []string a moran ga postavit
// na neku null vrijednost sta bi u ovon slucaju tribala bit ?nil?
func Discard() {
	SetOutput(ioutil.Discard)
	//golog.SetOutput(ioutil.Discard)
}

// Printf addes msg to log and writes log
// Problem dobijanja error elementa iz stringa saljen nil
func Printf(format string, v ...interface{}) {
	if !debugLogLevelEnabled {
		return
	}
	level, msg := splitLevelMessage(format)
	a := newAgregator(3)
	a.print(level, msg)
}

// Debug function receives message, sets: "level":"debug" and "msg":"message"
// in log buffer and then prints it
func Debug(msg string, v ...interface{}) {
	//logger.Debug(msg, v)
	newAgregator(4).Debug(sprintf(msg, v...))
}

// Info function receives message, sets: "level":"info" and "msg":"message"
// in log buffer and then prints it
func Info(msg string, v ...interface{}) {
	newAgregator(4).Info(sprintf(msg, v...))
}

// Error function receives error, sets: "level":"error" and "msg":"error"
// if error != nil or "msg":"" if error == nill
// in log buffer and then prints it
func Error(err error) {
	newAgregator(4).Error(err)
}

//Errorf function receives message, sets "level":"error" and "msg":"message"
//in log buffer and then prints it
func Errorf(msg string, v ...interface{}) {
	newAgregator(4).Error(fmt.Errorf(msg, v...))
}

// Notice function receives message, sets: "level":"notice" and "msg":"message"
// in log buffer and then prints it
func Notice(msg string, v ...interface{}) {
	newAgregator(4).Notice(sprintf(msg, v...))
}

// Fatal function receives error, sets: "level":"fatal" and "msg":"error"
// if error != nil or "msg":"" if error == nill
// in log buffer, prints buffer and then exits
func Fatal(err error) {
	newAgregator(4).Fatal(err)
}

// Fatalf function receives error, sets: "level":"fatal" and "msg":"error"
// in log buffer, prints buffer and then exits
func Fatalf(msg string, v ...interface{}) {
	newAgregator(4).Fatal(fmt.Errorf(msg, v...))
	os.Exit(-1)
}

//ako je v razlicit od 0 vrati sve inace vrati samo msg
func sprintf(msg string, v ...interface{}) string {
	if len(v) != 0 {
		return fmt.Sprintf(msg, v...)
	}
	return msg
}

//B finds and handles key-value pairs with boolean value
func B(key string, val bool) *Agregator {
	return newAgregator(3).B(key, val)
}

//I finds and handles key-value pairs with int value
func I(key string, val int) *Agregator {
	return newAgregator(1).I(key, val)
}

//F finds and handles key-value pairs with float64 value
func F(key string, val float64, prec int) *Agregator {
	return newAgregator(3).F(key, val, prec)
}

//S finds and handles key-value pairs with string value
func S(key string, val string) *Agregator {
	return newAgregator(3).S(key, val)
}

//J finds and handles key-value pairs with []byte value
func J(key string, val []byte) *Agregator {
	return newAgregator(3).J(key, val)
}

// Jc finds and handles key-value pairs with []byte value with check
func Jc(key string, val []byte) *Agregator {
	return newAgregator(3).Jc(key, val)
}

// Write sends byte array into output
// Nije implementirano u log2
func Write(buf []byte) {
	out.Write(buf)
}
