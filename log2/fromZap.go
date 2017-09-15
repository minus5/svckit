package log2

import (
	"os"
	"sort"
	"time"

	"go.uber.org/multierr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// build je implementacija klase Build iz zap.config
// rad je malo promjenjen ali radi isto
// koristi se zbog poziva interne funkcije openSinks
func build(cfg zap.Config, opts ...zap.Option) {
	enc := zapcore.NewJSONEncoder(cfg.EncoderConfig)

	sink, errSink, err := openSinks(cfg)
	if err != nil {
		return
	}
	logger := zap.New(
		zapcore.NewCore(enc, sink, cfg.Level),
		buildOptions(cfg, errSink)...,
	)

	if len(opts) > 0 {
		logger = logger.WithOptions(opts...)
	}

	a.zlog = logger
}

// openSinks je implementacija klase openSinks iz zap.config
// koristi se kako bi se mogla pozvati funkcija open
func openSinks(cfg zap.Config) (zapcore.WriteSyncer, zapcore.WriteSyncer, error) {
	sink, closeOut, err := open(cfg.OutputPaths...)
	if err != nil {
		return nil, nil, err
	}
	errSink, _, err := open(cfg.ErrorOutputPaths...)
	if err != nil {
		closeOut()
		return nil, nil, err
	}
	return sink, errSink, nil
}

// open je implementacija funkcije iz zap.write
// kopirana je kako bi se iz nje mogla pozvati interna klasa open
// implementirana pod nazivom o
func open(paths ...string) (zapcore.WriteSyncer, func(), error) {
	writers, close, err := o(paths)
	if err != nil {
		return nil, nil, err
	}
	writer := zap.CombineWriteSyncers(writers...)
	return writer, close, nil
}

// o je implementacija funkcije open iz zap.write
// koristi se kako bi se rucno moglo dodat pisanje sysloga
func o(paths []string) ([]zapcore.WriteSyncer, func(), error) {
	var openErr error
	writers := make([]zapcore.WriteSyncer, 0, len(paths))
	writers = append(writers, zapcore.AddSync(out))
	files := make([]*os.File, 0, len(paths))
	close := func() {
		for _, f := range files {
			f.Close()
		}
	}
	for _, path := range paths {
		switch path {
		case "stdout":
			writers = append(writers, os.Stdout)
			// Don't close standard out.
			continue
		case "stderr":
			writers = append(writers, os.Stderr)
			// Don't close standard error.
			continue
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		openErr = multierr.Append(openErr, err)
		if err == nil {
			writers = append(writers, f)
			files = append(files, f)
		}
	}

	if openErr != nil {
		close()
		return writers, nil, openErr
	}

	return writers, close, nil
}

// buildOptions je kopirana funkcija iz zap.config
// kopirana je jer je interna a potrebno je promjeniti nacin stvaranja logera
// pa se iz tog razloga nemoze pozvati
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
