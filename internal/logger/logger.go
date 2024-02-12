package logger

import (
	"github.com/ids79/anti-bruteforce/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logg interface {
	Error(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
}

func New(config config.LoggerConf, servis string) *zap.SugaredLogger {
	lavel, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return nil
	}
	logConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(lavel),
		DisableCaller:     true,
		Development:       true,
		DisableStacktrace: true,
		Encoding:          config.LogEncoding,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
	}
	logger := zap.Must(logConfig.Build()).Sugar().Named(servis)
	return logger
}
