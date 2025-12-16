package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func Init(level, format string) error {
	cfg := zap.NewDevelopmentConfig()

	if format == "console" {
		cfg.Encoding = "console"
	}

	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return err
	}

	cfg.Level = zap.NewAtomicLevelAt(lvl)
	logger, err := cfg.Build()
	if err != nil {
		return err
	}

	log = logger
	return nil
}

func L() *zap.Logger {
	if log == nil {
		return zap.NewNop()
	}
	return log
}
