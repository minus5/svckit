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

	writers := make([]zapcore.WriteSyncer, 0)
	writers = append(writers, zapcore.AddSync(out))
	writer := zap.CombineWriteSyncers(writers...)

	logger := zap.New(
		zapcore.NewCore(enc, writer, cfg.Level),
		buildOptions(cfg, writer)...,
	)

	if len(opts) > 0 {
		logger = logger.WithOptions(opts...)
	}

	return logger
}

/*
// openSinks je implementacija klase openSinks iz zap.config
// koristi se kako bi se mogla pozvati funkcija open
func openSinks(cfg zap.Config) (zapcore.WriteSyncer, zapcore.WriteSyncer, error) {
	//sink, closeOut, err := open()
	sink, _, err := open()
	if err != nil {
		return nil, nil, err
	}
	errSink, _, err := open()
	if err != nil {
		//closeOut()
		return nil, nil, err
	}
	return sink, errSink, nil
}

// open je implementacija funkcije iz zap.write
// kopirana je kako bi se iz nje mogla pozvati interna klasa open
// implementirana pod nazivom o
func open() (zapcore.WriteSyncer, *func(), error) {
	writers, close, err := o()
	if err != nil {
		return nil, nil, err
	}
	writer := zap.CombineWriteSyncers(writers...)
	return writer, close, nil
}

// o je implementacija funkcije open iz zap.write
// koristi se kako bi se rucno moglo dodat pisanje sysloga
func o() ([]zapcore.WriteSyncer, *func(), error) {
	var openErr error
	writers := make([]zapcore.WriteSyncer, 0)
	files := make([]*os.File, 0)
	close := func() {
		for _, f := range files {
			f.Close()
		}
	}

	writers = append(writers, zapcore.AddSync(out))

	//f, err := os.OpenFile("stderr", os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	//openErr = multierr.Append(openErr, err)
	//if err == nil {
	//writers = append(writers, f)
	//files = append(files, f)
	//}

	if openErr != nil {
		close()
		//return nil, writers, openErr
		return writers, nil, openErr
	}

	//return nil, writers, nil
	return writers, nil, nil
}
*/

// buildOptions is function from zap library
// it is copied so it can be called from build function
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
