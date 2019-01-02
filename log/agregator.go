package log

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Agregator - spaja json log liniju.
// Po uzoru na kod u https://golang.org/src/log/log.go ne koristi fmt.Sprintf ni bytes.Buffer
// nego dodaje na neki postojeci buffer. Odatle vecina komplikacije.
// Rezultat je da je brz otprilike kao i logger iz standard go lib-a.
// Trenutno podrzava samo int i string tipove.
type Agregator struct {
	buf         *[]byte
	t           time.Time
	file        string
	line        int
	attrs       []*attr
	level       string
	msg         string
	callerDepth int
	output      io.Writer
}

const (
	// syslog udp paket ima limit
	// iskusveno sam zapeo u korisnom dijelu od 8047-8055 znakova
	// pa stavljam limit na pojedinacni string:
	MaxStrLen = 7500
)

type attr struct {
	key string
	val string
}

var bufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0)
		return &buf
	},
}

func NewAgregator(output io.Writer, callerDepth int) *Agregator {
	if output == nil {
		output = out
	}
	return &Agregator{
		buf:         nil,
		t:           time.Now(),
		callerDepth: callerDepth,
		output:      output,
	}
}

func newAgregator(callerDepth int) *Agregator {
	return &Agregator{
		buf:         nil,
		t:           time.Now(),
		callerDepth: callerDepth,
		output:      out,
	}
}

//get buffer from pool
func (a *Agregator) getBuf() {
	a.buf = bufPool.Get().(*[]byte)
}

//shrink and return to pool
func (a *Agregator) freeBuf() {
	buf := *a.buf
	buf = buf[:0]
	bufPool.Put(&buf)
}

func (a *Agregator) write() error {
	if a.file == "" { //zbog testova
		a.file, a.line = getCaller(a.callerDepth)
	}
	a.msg = limitStrLen(strconv.QuoteToASCII(a.msg))
	a.getBuf()
	a.timeFile(a.t, a.file, a.line)
	a.s("level", a.level)
	for _, atr := range a.attrs {
		a.s(atr.key, atr.val)
	}
	a.s("msg", a.msg)
	*a.buf = append(*a.buf, "}\n"...)
	_, err := a.output.Write(*a.buf)
	a.freeBuf()
	return err
}

func (a *Agregator) timeFile(t time.Time, file string, line int) {
	buf := a.buf
	*buf = append(*buf, `{"time":"`...)
	year, month, day := t.Date()
	itoa(buf, year, 4)
	*buf = append(*buf, '-')
	itoa(buf, int(month), 2)
	*buf = append(*buf, '-')
	itoa(buf, day, 2)
	*buf = append(*buf, 'T')
	hour, min, sec := t.Clock()
	itoa(buf, hour, 2)
	*buf = append(*buf, ':')
	itoa(buf, min, 2)
	*buf = append(*buf, ':')
	itoa(buf, sec, 2)
	*buf = append(*buf, '.')
	itoa(buf, t.Nanosecond()/1e3, 6)
	//itoa(buf, t.Nanosecond(), 9)
	//TODO - ovo uzima u obzir samo zone po puni sat
	_, offset := t.Zone()
	*buf = append(*buf, '+')
	itoa(buf, offset/3600, 2)
	*buf = append(*buf, `:00`...)

	*buf = append(*buf, `", "file":"`...)
	*buf = append(*buf, file...)
	*buf = append(*buf, ':')
	itoa(buf, line, -1)
	if len(prefix) > 0 {
		*buf = append(*buf, `", `...)
		*buf = append(*buf, prefix...)
	} else {
		*buf = append(*buf, `"`...)
	}
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

var (
	reservedKeys = []string{"host", "app", "level", "msg", "file", "time"}
)

func escapeKey(key string) string {
	for _, k := range reservedKeys {
		if key == k {
			return "_" + key
		}
	}
	return key
}

func (a *Agregator) Debug(msg string) {
	if !debugLogLevelEnabled {
		return
	}
	a.level = LevelDebug
	a.msg = msg
	a.write()
}

func (a *Agregator) Info(msg string) {
	a.level = LevelInfo
	a.msg = msg
	a.write()
}

func (a *Agregator) ErrorS(msg string) {
	a.level = LevelError
	a.msg = msg
	a.write()
}

func (a *Agregator) Error(err error) {
	a.level = LevelError
	if err != nil {
		a.msg = err.Error()
	} else {
		a.msg = ""
	}
	a.write()
}

func (a *Agregator) Notice(msg string) {
	a.level = LevelNotice
	a.msg = msg
	a.write()
}

func (a *Agregator) Event(msg string) {
	a.level = LevelEvent
	a.msg = msg
	a.write()
}

func (a *Agregator) Fatal(err error) {
	a.level = LevelFatal
	if err != nil {
		a.msg = err.Error()
	} else {
		a.msg = ""
	}
	a.write()
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
	a.attrs = append(a.attrs, &attr{key: key, val: fmt.Sprintf("%v", val)})
	return a
}

// I - add integer key, value attribute
func (a *Agregator) I(key string, val int) *Agregator {
	key = escapeKey(key)
	s := strconv.Itoa(val)
	a.attrs = append(a.attrs, &attr{key: key, val: s})
	return a
}

// F - add float64 key, value attribute
func (a *Agregator) F(key string, val float64, prec int) *Agregator {
	key = escapeKey(key)
	s := strconv.FormatFloat(val, 'f', prec, 64)
	a.attrs = append(a.attrs, &attr{key: key, val: s})
	return a
}

// S - add string key, value attribute
func (a *Agregator) S(key string, val string) *Agregator {
	key = escapeKey(key)
	val = limitStrLen(strconv.QuoteToASCII(val))
	a.attrs = append(a.attrs, &attr{key: key, val: val})
	return a
}

// Add to buffer key and escaped string value
func (a *Agregator) s(key string, val string) *Agregator {
	*a.buf = append(*a.buf, ',')
	*a.buf = append(*a.buf, ' ')
	*a.buf = append(*a.buf, '"')
	*a.buf = append(*a.buf, key...)
	*a.buf = append(*a.buf, '"')
	*a.buf = append(*a.buf, ':')
	*a.buf = append(*a.buf, val...)
	return a
}

// Dodaj json atribut.
// Odgovornost je aplikacije da je val validan json.
func (a *Agregator) J(key string, val []byte) *Agregator {
	key = escapeKey(key)
	if val == nil || len(val) == 0 {
		a.attrs = append(a.attrs, &attr{key: key, val: "null"})
		return a
	}
	if len(val) > MaxStrLen {
		return a.S(key, string(val))
	}
	a.attrs = append(a.attrs, &attr{key: key, val: string(val)})
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

// Cheap integer to fixed-width decimal ASCII.  Give a negative width to avoid zero-padding.
// Only for positive ints.
func itoa(buf *[]byte, i int, wid int) {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	*buf = append(*buf, b[bp:]...)
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

func getCaller(depth int) (string, int) {
	_, longFile, line, ok := runtime.Caller(depth)
	if !ok {
		longFile = "???"
		line = 0
	}
	shortFile := longFile
	for i := len(longFile) - 1; i > 0; i-- {
		if longFile[i] == '/' {
			shortFile = longFile[i+1:]
			break
		}
	}
	return shortFile, line
}
