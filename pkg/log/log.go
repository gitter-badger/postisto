package log

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _log *zap.SugaredLogger

func init() {
	var logConfig  config.LogConfig
	logConfig.PreSetMode = "debug"

	if err := InitWithConfig(logConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %q\n\nThat should not happen. Please call a doctor.", err.Error())) //TODO bad idea?
	}
}

func InitWithConfig(logConfig config.LogConfig) error {
	if _log != nil {
		Debug("Logger already initialized. Will re-initialize now..")
	}

	var cfg *zap.Config

	switch logConfig.PreSetMode {
	case "debug":
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		cfg = &config
	case "dev":
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg = &config
	case "prod":
		config := zap.NewProductionConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg = &config
	default:
		if logConfig.ZapConfig == nil {
			return fmt.Errorf("log config not set")
		}

		cfg = logConfig.ZapConfig
	}

	rawLogger, err := cfg.Build()
	defer rawLogger.Sync()

	if err != nil {
		return err
	}

	_log = rawLogger.WithOptions(zap.AddCallerSkip(1)).Sugar() // pkg variable
	Info("logging successfully initialized")

	return err
}

func Error(msg string) {
	_log.Error(msg)
}

func Errorw(msg string,  context ...interface{}) {
	_log.With(context...).Error(msg)
}

func Info(msg string) {
	_log.Info(msg)
}

func Infow(msg string, context ...interface{}) {
	_log.With(context...).Info(msg)
}

func Debug(msg string) {
	_log.Debug(msg)
}

func Debugw(msg string,  context ...interface{}) {
	_log.With(context...).Debug(msg)
}
