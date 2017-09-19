package log2

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/minus5/svckit/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Agregator - spaja json log liniju.
// Po uzoru na kod u https://golang.org/src/log/log.go ne koristi fmt.Sprintf ni bytes.Buffer
// nego dodaje na neki postojeci buffer. Odatle vecina komplikacije.
// Rezultat je da je brz otprilike kao i logger iz standard go lib-a.
// Trenutno podrzava samo int i string tipove.
type Agregator struct {
	zlog         *zap.Logger
	commonFields []zapcore.Field
	fields       []zapcore.Field
}

const (
	// syslog udp paket ima limit
	// iskusveno sam zapeo u korisnom dijelu od 8047-8055 znakova
	// pa stavljam limit na pojedinacni string:
	MaxStrLen = 7500
)

// isto ka pool u library-ju
var (
//a Agregator
)

// return as quoted string
var (
	LevelDebug  = `"debug"`
	LevelInfo   = `"info"`
	LevelError  = `"error"`
	LevelFatal  = `"fatal"`
	LevelNotice = `"notice"`
	LevelEvent  = `"event"`
)

// unquoted versions
var (
	LevelDebugUnquoted  = strings.Trim(LevelDebug, `"`)
	LevelInfoUnquoted   = strings.Trim(LevelInfo, `"`)
	LevelErrorUnquoted  = strings.Trim(LevelError, `"`)
	LevelFatalUnquoted  = strings.Trim(LevelFatal, `"`)
	LevelNoticeUnquoted = strings.Trim(LevelNotice, `"`)
	LevelEventUnquoted  = strings.Trim(LevelEvent, `"`)
)

// reserved keys
var (
	reservedKeys = []string{"host", "app", "level", "msg", "file", "time"}
)

//NewAgregator creates new agregator based on output writer and depth
func NewAgregator(output io.Writer, depth int) *Agregator {
	if output == nil {
		output = out
	}
	loger := build(cfg, out, zap.Fields(
		zap.String("host", env.Hostname()),
		zap.String("app", env.AppName()),
	), zap.AddCallerSkip(depth))

	a := &Agregator{
		zlog: loger,
	}
	return a
}

//newAgregator creates new agregator based on depth
func newAgregator(depth int) *Agregator {
	agregator := &Agregator{
		zlog: a.zlog.WithOptions(zap.AddCallerSkip(depth)),
	}
	return agregator
}

// New function creates new agregator
func New() *Agregator {
	return newAgregator(1).Build()
}

// Build function adds fields as commonFields
func (a *Agregator) Build() *Agregator {
	a = &Agregator{
		commonFields: a.fields,
		zlog:         a.zlog,
	}
	return a
}

// New function adds fields as commonFields
func (a *Agregator) New() *Agregator {
	return a.Build()
}

//print switches printing between different levels
func (a *Agregator) print(level, msg string) {
	switch level {
	case "debug":
		a.Debug(msg)
	case "info":
		a.Info(msg)
	case "notice":
		a.Notice(msg)
	case "errors":
		a.ErrorS(msg)
	case "error":
		a.Error(nil)
	case "fatal":
		a.Fatal(nil)
	case "event":
		a.Event(msg)
	}
}

//escapeKey assures no additional key matches system keys
func escapeKey(key string) string {
	for _, k := range reservedKeys {
		if key == k {
			return "_" + key
		}
	}
	return key
}

// Debug function prints log with "level":"debug"
func (a *Agregator) Debug(msg string) {
	a.zlog.Debug(msg, append(a.fields, a.commonFields...)...)
	a.Sync()
}

// Info function prints log with "level":"info"
func (a *Agregator) Info(msg string) {
	a.zlog.Info(msg, append(a.fields, a.commonFields...)...)
	a.Sync()
}

// ErrorS function prints log with "level":"error"
func (a *Agregator) ErrorS(msg string) {
	a.zlog.Error(msg, append(a.fields, a.commonFields...)...)
	a.Sync()
}

// Error function prints log with "level":"error"
func (a *Agregator) Error(err error) {
	var msg string
	if err != nil {
		msg = err.Error()
	} else {
		msg = ""
	}
	a.ErrorS(msg)
}

// Notice function prints log with "level":"info" and "notice":"info"
func (a *Agregator) Notice(msg string) {
	a.fields = append(a.fields, zap.String("notice", "info"))
	a.zlog.Info(msg, append(a.fields, a.commonFields...)...)
	a.Sync()

	//ce := *a.zlog.Check(zap.InfoLevel, msg)
	//ce.Write()
}

// Event function prints log with "level":"info" and "event":"info"
func (a *Agregator) Event(msg string) {
	a.fields = append(a.fields, zap.String("event", "info"))
	a.zlog.Info(msg, append(a.fields, a.commonFields...)...)
	a.Sync()
}

// Fatal function prints log with "level":"fatal"
func (a *Agregator) Fatal(err error) {
	var msg string
	if err != nil {
		msg = err.Error()
	} else {
		msg = ""
	}
	a.zlog.Fatal(msg, a.fields...)
	a.Sync()
	a.fields = nil
	os.Exit(-1)
}

// Sync function syncs logger
func (a *Agregator) Sync() {
	a.zlog.Sync()
}

// udp syslog poruka ima limit ~8k
// radi sa stringom koji je dobiven nakon QuoteToASCII (ima " na pocetku i kraju)
func limitStrLen(s string) string {
	if len(s) <= MaxStrLen {
		return s
	}
	s = s[:MaxStrLen-4]
	// ako si odrezao u sred quote izvrti do pocetaka
	for s[len(s)-1:] == "\\" {
		s = s[:len(s)-1]
	}
	return s + `..."`
}

// B - add boolean key:value attribute
func (a *Agregator) B(key string, val bool) *Agregator {
	if a.fields == nil {
		a = &Agregator{
			zlog:         a.zlog,
			commonFields: a.commonFields,
		}
	}

	key = escapeKey(key)
	//a.zlog = a.zlog.With(zap.Bool(key, val))
	a.fields = append(a.fields, zap.Bool(key, val))
	return a
}

// I - add integer key:value attribute
func (a *Agregator) I(key string, val int) *Agregator {
	if a.fields == nil {
		a = &Agregator{
			zlog:         a.zlog,
			commonFields: a.commonFields,
		}
	}

	key = escapeKey(key)
	//a.zlog = a.zlog.With(zap.Int(key, val))
	a.fields = append(a.fields, zap.Int(key, val))
	return a
}

// F - add float64 key:value attribute
func (a *Agregator) F(key string, val float64, prec int) *Agregator {
	if a.fields == nil {
		a = &Agregator{
			zlog:         a.zlog,
			commonFields: a.commonFields,
		}
	}

	key = escapeKey(key)
	s := strconv.FormatFloat(val, 'f', prec, 64)
	a.fields = append(a.fields, zap.String(key, s))
	return a
}

// S - add string key:value attribute
func (a *Agregator) S(key string, val string) *Agregator {
	if a.fields == nil {
		a = &Agregator{
			zlog:         a.zlog,
			commonFields: a.commonFields,
		}
	}

	key = escapeKey(key)
	if len(val) > MaxStrLen {
		val = limitStrLen(strconv.QuoteToASCII(val))
	}
	//a.zlog = a.zlog.With(zap.String(key, val))
	a.fields = append(a.fields, zap.String(key, val))
	return a
}

// J - add json key:value attribute
// It is applications responsibility to asure valid json
func (a *Agregator) J(key string, val []byte) *Agregator {
	if a.fields == nil {
		a = &Agregator{
			zlog:         a.zlog,
			commonFields: a.commonFields,
		}
	}

	key = escapeKey(key)
	if val == nil || len(val) == 0 {
		//a.zlog = a.zlog.With(zap.String(key, "null"))
		a.fields = append(a.fields, zap.String(key, "null"))
		return a
	}
	if len(val) > MaxStrLen {
		return a.S(key, string(val))
	}
	//a.zlog = a.zlog.With(zap.String(key, string(val)))
	a.fields = append(a.fields, zap.String(key, string(val)))
	return a
}

// Jc - add json key:value attribute
// Contains control
func (a *Agregator) Jc(key string, val []byte) *Agregator {
	key = escapeKey(key)
	var m map[string]interface{}
	err := json.Unmarshal(val, &m)
	if err == nil {
		return a.J(key, val)
	}
	return a.S(key, string(val))
}

//splitLevelMessage cuts key words from msg
func splitLevelMessage(line string) (string, string) {
	if !strings.Contains(line, "[") {
		if strings.Contains(line, "error") {
			return LevelError, line
		}
		return LevelDebug, line
	}
	replace := func(s, old string) string {
		return strings.TrimPrefix(strings.Replace(s, old, "", 1), " ")
	}
	if strings.Contains(line, "[DEBUG]") {
		return LevelDebug, replace(line, "[DEBUG]")
	}
	if strings.Contains(line, "[INFO]") {
		return LevelInfo, replace(line, "[INFO]")
	}
	if strings.Contains(line, "[ERROR]") {
		return LevelError, replace(line, "[ERROR]")
	}
	if strings.Contains(line, "[NOTICE]") {
		return LevelNotice, replace(line, "[NOTICE]")
	}
	return LevelDebug, line
}
