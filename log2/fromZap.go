package log2

import (
	"io"
	"sort"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// build is implementation of Build from zap.config
// it is implemented to write on output of choice
func build(cfg zap.Config, out io.Writer, opts ...zap.Option) *zap.Logger {
	enc := zapcore.NewJSONEncoder(cfg.EncoderConfig)

	writer := zapcore.AddSync(out)

	logger := zap.New(
		zapcore.NewCore(enc, writer, cfg.Level),
		buildOptions(cfg, writer)...,
	)

	if len(opts) > 0 {
		logger = logger.WithOptions(opts...)
	}

	return logger
}

func buildOptions(cfg zap.Config, errSink zapcore.WriteSyncer) []zap.Option {
	opts := []zap.Option{zap.ErrorOutput(errSink)}

	if !cfg.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}

	stackLevel := zapcore.ErrorLevel
	if cfg.Development {
		stackLevel = zapcore.WarnLevel
	}
	if !cfg.DisableStacktrace {
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}

	if cfg.Sampling != nil {
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(core, time.Second, int(cfg.Sampling.Initial), int(cfg.Sampling.Thereafter))
		}))
	}

	if len(cfg.InitialFields) > 0 {
		fs := make([]zapcore.Field, 0, len(cfg.InitialFields))
		keys := make([]string, 0, len(cfg.InitialFields))
		for k := range cfg.InitialFields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fs = append(fs, zap.Any(k, cfg.InitialFields[k]))
		}
		opts = append(opts, zap.Fields(fs...))
	}

	return opts
}
