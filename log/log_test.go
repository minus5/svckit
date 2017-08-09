/*
While running benchmark redirect stderr to /dev/null

  go test -bench=.  2> /dev/null
*/
package log

import (
	"os"
	"runtime"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/minus5/svckit/env"
	"github.com/stretchr/testify/assert"

	golog "log"
)

func BenchmarkStandardLogger(b *testing.B) {
	std := golog.New(os.Stderr, "", golog.LstdFlags|golog.Lshortfile|golog.Lmicroseconds)
	for n := 0; n < b.N; n++ {
		std.Printf("time, file, level: DEBUG iso medo u ducan %d puta, pero %s, key: %s msg: 'iso medo u ducan'", n, "zdero", "value")
	}
}

func BenchmarkSvckitLog(b *testing.B) {
	for n := 0; n < b.N; n++ {
		I("puta", n).F("float64", 3.1415926535, -1).S("pero", "zdero").S("key", "value").Debug("iso medo u ducan")
	}
}

func BenchmarkSvckitLogPrintf(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Printf("[INFO] iso medo u ducan %d puta, pero %s, key: %s ", n, "zdero", "value")
	}
}

func BenchmarkSvckitLogPrintfBezLevel(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Printf("iso medo u ducan %d puta, pero %s, key: %s ", n, "zdero", "value")
	}
}

func BenchmarkSvckitLogPrintfBezLevelBezFmt(b *testing.B) {
	for n := 0; n < b.N; n++ {
		Printf("iso medo u ducan x puta")
	}
}

// //usporedba s popularim structured logger-om
// func BenchmarkStructuredLogrus(b *testing.B) {
// 	logrus.SetFormatter(&logrus.JSONFormatter{})
// 	for n := 0; n < b.N; n++ {
// 		logrus.WithFields(logrus.Fields{
// 			"puta":  n,
// 			"app":   "app",
// 			"host":  "host",
// 			"level": "debug",
// 			"time":  time.Now(),
// 			"no":    n,
// 		}).Info("iso medo u ducan")
// 	}
// }
/*
BenchmarkStandardLogger-4               	  500000	      2384 ns/op
BenchmarkSvckitLog-4                    	  300000	      3666 ns/op
BenchmarkSvckitLogPrintf-4              	  300000	      3509 ns/op
BenchmarkSvckitLogPrintfBezLevel-4      	  500000	      3048 ns/op
BenchmarkSvckitLogPrintfBezLevelBezFmt-4	 1000000	      2267 ns/op
BenchmarkStructuredLogrus-4             	  100000	     12627 ns/op
*/

/* razlika replace i regexp implementacije:
BenchmarkSplitLevelMsgRegexp-4   	  300000	      3982 ns/op
BenchmarkSplitLevelMsgAlternate-4	 2000000	       601 ns/op
*/

func TestZap(t *testing.T) {
	n := 1
	I("puta", n).F("float64", 3.1415926535, -1).S("pero", "zdero").S("key", "value").Info("iso medo u ducan")
	//logger, _ := zap.NewProduction()
	//logger := zap.New(zapcore.)
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.CallerKey = "file"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := config.Build(zap.Fields(
		zap.String("host", env.Hostname()),
		zap.String("app", env.AppName()),
	))
	defer logger.Sync()
	logger.Info("iso medo u ducan",
		zap.Int("puta", n),
		zap.Float64("float64", 3.1415926535),
		zap.String("pero", "zdero"),
		zap.String("key", "value"),
	)
}

func BenchmarkZap(b *testing.B) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "time"
	config.EncoderConfig.CallerKey = "file"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ := config.Build(zap.Fields(
		zap.String("host", env.Hostname()),
		zap.String("app", env.AppName()),
	))
	defer logger.Sync()
	//sugar := logger.Sugar()

	for n := 0; n < b.N; n++ {
		logger.Info("iso medo u ducan",
			zap.Int("puta", n),
			zap.Float64("float64", 3.1415926535),
			zap.String("pero", "zdero"),
			zap.String("key", "value"),
		)
	}
}

func BenchmarkZapSvckit(b *testing.B) {
	for n := 0; n < b.N; n++ {
		I("puta", n).F("float64", 3.1415926535, -1).S("pero", "zdero").S("key", "value").Debug("iso medo u ducan")
	}
}

func BenchmarkSplitLevelMsg(b *testing.B) {
	line := "[NOTICE] neki message koji ide nakon toga"
	for n := 0; n < b.N; n++ {
		splitLevelMessage(line)
	}
}

func TestSplitLevelMessage(t *testing.T) {
	data := []struct {
		line  string
		level string
		msg   string
	}{
		{"[DEBUG] nesto", LevelDebug, "nesto"},
		{"[NOTICE] nesto", LevelNotice, "nesto"},
		{"[ERROR] nesto", LevelError, "nesto"},
		{"error nesto", LevelError, "error nesto"},
		{"pero nesto", LevelDebug, "pero nesto"},
	}

	for _, d := range data {
		level, msg := splitLevelMessage(d.line)
		assert.Equal(t, d.level, level)
		assert.Equal(t, d.msg, msg)
	}
}

func BenchmarkRuntimeCallerDepth4(b *testing.B) {
	for n := 0; n < b.N; n++ {
		runtime.Caller(4)
	}
}

func BenchmarkRuntimeCallerDepth3(b *testing.B) {
	for n := 0; n < b.N; n++ {
		runtime.Caller(3)
	}
}

func BenchmarkRuntimeCallerDepth2(b *testing.B) {
	for n := 0; n < b.N; n++ {
		runtime.Caller(2)
	}
}

func BenchmarkRuntimeCallerDepth1(b *testing.B) {
	for n := 0; n < b.N; n++ {
		runtime.Caller(1)
	}
}
