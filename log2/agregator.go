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

// Agregator structure is used to save all data of single log output
// zlog field is used to save zapLogger structure
// commonFields is array of all key:value pairs used in more than one output of same instance of logger (type: zapcore.Field)
// fields is array of all key:value pairs used in single output by instance of logger (type: zapcore.Field)
type Agregator struct {
	zlog         *zap.Logger
	commonFields []zapcore.Field
	fields       []zapcore.Field
}

const (
	// MaxStrLen assures that length of string does not pass 7500 characters
	// syslog udp package has limit (in experience no more than 8047-8055 usefull characters)
	MaxStrLen = 7500
)

// return as quoted string
var (
	LevelDebug  = `debug`
	LevelInfo   = `info`
	LevelError  = `error`
	LevelFatal  = `fatal`
	LevelNotice = `notice`
	LevelEvent  = `event`
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

	loger := build(cfg, output, zap.Fields(
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

// ClearFields function sets fields array as nil
func (a *Agregator) ClearFields() *Agregator {
	a = &Agregator{
		commonFields: a.commonFields,
		zlog:         a.zlog,
	}
	return a
}

// ClearCommonFields function sets commonFields array as nil
func (a *Agregator) ClearCommonFields() *Agregator {
	a = &Agregator{
		fields: a.fields,
		
		zlog:   a.zlog,
	}
	return a
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
	default:
		a.Debug(msg)
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
	a.fields = nil
}

// Info function prints log with "level":"info"
func (a *Agregator) Info(msg string) {
	a.zlog.Info(msg, append(a.fields, a.commonFields...)...)
	a.fields = nil
}

// ErrorS function prints log with "level":"error"
func (a *Agregator) ErrorS(msg string) {
	a.zlog.Error(msg, append(a.fields, a.commonFields...)...)
	a.fields = nil
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
	a.fields = nil
}

// Event function prints log with "level":"info" and "event":"info"
func (a *Agregator) Event(msg string) {
	a.fields = append(a.fields, zap.String("event", "info"))
	a.zlog.Info(msg, append(a.fields, a.commonFields...)...)
	a.fields = nil
}

// Fatal function prints log with "level":"fatal"
func (a *Agregator) Fatal(err error) {
	var msg string
	if err != nil {
		msg = err.Error()
	} else {
		msg = ""
	}
	a.zlog.Fatal(msg, append(a.fields, a.commonFields...)...)
	a.fields = nil
	os.Exit(-1)
}

// Sync function syncs logger
func (a *Agregator) Sync() {
	a.zlog.Sync()
}

// udp syslog message has limit ~8k
// use string after QuoteToASCII (" on the start and at the end)
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
	key = escapeKey(key)

	a.fields = append(a.fields, zap.Bool(key, val))
	return a
}

// I - add integer key:value attribute
func (a *Agregator) I(key string, val int) *Agregator {
	key = escapeKey(key)

	a.fields = append(a.fields, zap.Int(key, val))
	return a
}

// F - add float64 key:value attribute
func (a *Agregator) F(key string, val float64, prec int) *Agregator {
	key = escapeKey(key)
	a.fields = append(a.fields, zap.Float64(key, val))
	return a
}

// S - add string key:value attribute
func (a *Agregator) S(key string, val string) *Agregator {
	key = escapeKey(key)
	if len(val) > MaxStrLen {
		val = limitStrLen(strconv.QuoteToASCII(val))
	}

	a.fields = append(a.fields, zap.String(key, val))
	return a
}

// J - add json key:value attribute
// It is applications responsibility to assure valid json
func (a *Agregator) J(key string, val []byte) *Agregator {
	key = escapeKey(key)
	if val == nil || len(val) == 0 {
		a.fields = append(a.fields, zap.String(key, "null"))
		return a
	}
	if len(val) > MaxStrLen {
		return a.S(key, string(val))
	}

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
