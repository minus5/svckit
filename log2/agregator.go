package log2

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/minus5/svckit/env"
	"go.uber.org/zap"
)

// Agregator - spaja json log liniju.
// Po uzoru na kod u https://golang.org/src/log/log.go ne koristi fmt.Sprintf ni bytes.Buffer
// nego dodaje na neki postojeci buffer. Odatle vecina komplikacije.
// Rezultat je da je brz otprilike kao i logger iz standard go lib-a.
// Trenutno podrzava samo int i string tipove.
type Agregator struct {
	zlog zap.Logger
	//zcon zap.Config
}

const (
	// syslog udp paket ima limit
	// iskusveno sam zapeo u korisnom dijelu od 8047-8055 znakova
	// pa stavljam limit na pojedinacni string:
	MaxStrLen = 7500
)

// isto ka pool u library-ju
var bufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0)
		return &buf
	},
}

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

func newAgregator(depth int) *Agregator {
	a := Agregator{}
	//a.zcon = zap.NewDevelopmentConfig()
	//a.zcon.EncoderConfig.TimeKey = "time"
	//a.zcon.EncoderConfig.CallerKey = "file"
	//a.zcon.EncoderConfig.LevelKey = "level"
	//a.zcon.EncoderConfig.MessageKey = "msg"
	//a.zcon.EncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	//a.zcon.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	//a.zcon.Development = false
	//a.zcon.Encoding = "json"
	logger, _ := cfg.Build(zap.Fields(
		zap.String("host", env.Hostname()),
		zap.String("app", env.AppName()),
	), zap.AddCallerSkip(depth))
	a.zlog = *logger
	//defer a.zlog.Sync()

	return &a
}

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

func escapeKey(key string) string {
	for _, k := range reservedKeys {
		if key == k {
			return "_" + key
		}
	}
	return key
}

func (a *Agregator) Debug(msg string) {
	a.zlog.Debug(msg)
}

func (a *Agregator) Info(msg string) {
	a.zlog.Info(msg)
}

func (a *Agregator) ErrorS(msg string) {
	a.zlog.Error(msg)
}

func (a *Agregator) Error(err error) {
	var msg string
	if err != nil {
		msg = err.Error()
	} else {
		msg = ""
	}
	a.ErrorS(msg)
}

// NEMA NOTICE-A U STANDARDNOM ZAP paketu
func (a *Agregator) Notice(msg string) {
	a.zlog = *a.zlog.With(zap.String("notice", "info"))
	a.zlog.Info(msg)

	//ce := *a.zlog.Check(zap.InfoLevel, msg)
	//ce.Write()
}

// NEMA EVENT-A U STANDARDNOM ZAP PAKETU
func (a *Agregator) Event(msg string) {
	a.zlog = *a.zlog.With(zap.String("event", "info"))
	a.zlog.Info(msg)
}

func (a *Agregator) Fatal(err error) {
	var msg string
	if err != nil {
		msg = err.Error()
	} else {
		msg = ""
	}
	a.zlog.Fatal(msg)
	os.Exit(-1)
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

// B - add boolean key, value attribute
func (a *Agregator) B(key string, val bool) *Agregator {
	key = escapeKey(key)
	a.zlog = *a.zlog.With(zap.Bool(key, val))
	return a
}

// I - add integer key, value attribute
func (a *Agregator) I(key string, val int) *Agregator {
	key = escapeKey(key)
	a.zlog = *a.zlog.With(zap.Int(key, val))
	return a
}

// F - add float64 key, value attribute
func (a *Agregator) F(key string, val float64, prec int) *Agregator {
	key = escapeKey(key)
	//s := strconv.FormatFloat(val, 'f', prec, 64)
	a.zlog = *a.zlog.With(zap.Float64(key, val))
	return a
}

// S - add integer key, value attribute
// val se umisto "val" ispisuje "\"val\"" !!!!!!!
func (a *Agregator) S(key string, val string) *Agregator {
	key = escapeKey(key)
	if len(val) > MaxStrLen {
		val = limitStrLen(strconv.QuoteToASCII(val))
	}
	a.zlog = *a.zlog.With(zap.String(key, val))

	return a
}

// Dodaj json atribut.
// Odgovornost je aplikacije da je val validan json.
// Nisan siguran triba li se dodavat priko stringa ili priko byte-a
func (a *Agregator) J(key string, val []byte) *Agregator {
	key = escapeKey(key)
	if val == nil || len(val) == 0 {
		a.zlog = *a.zlog.With(zap.String(key, "null"))
		return a
	}
	if len(val) > MaxStrLen {
		return a.S(key, string(val))
	}
	a.zlog = *a.zlog.With(zap.String(key, string(val)))
	return a
}

// Isto kao j ali provjerava da li je val validan json
func (a *Agregator) Jc(key string, val []byte) *Agregator {
	key = escapeKey(key)
	var m map[string]interface{}
	err := json.Unmarshal(val, &m)
	if err == nil {
		return a.J(key, val)
	}
	return a.S(key, string(val))
}

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
