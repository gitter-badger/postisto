package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

type Config struct {
	PreSetMode string      `yaml:"mode"`
	ZapConfig  *zap.Config `yaml:"dangerousAdvancedZapConfig"`
}

func init() {
	var logConfig Config
	logConfig.PreSetMode = "debug"

	if err := InitWithConfig(logConfig); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %q\n\nThat should not happen. Please call a doctor.", err.Error()))
	}
}

func InitWithConfig(logConfig Config) error {
	if log != nil {
		Debug("Logger already initialized. Will re-initialize now..")
	}

	var cfg *zap.Config

	switch logConfig.PreSetMode {
	case "debug":
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		config.DisableStacktrace = true
		cfg = &config
	case "dev":
		config := zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.DisableStacktrace = true
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

	if err != nil {
		return err
	}
	defer rawLogger.Sync()

	log = rawLogger.WithOptions(zap.AddCallerSkip(1)).Sugar() // pkg variable
	Info("logging successfully initialized")

	return err
}

func Panic(msg string, err error) {
	log.With("err", err).Panic(msg)
}

func Panicw(msg string, err error, context ...interface{}) {
	context = append(context, "err")
	context = append(context, err)
	log.With(context...).Panic(msg)
}

func Error(msg string, err error) {
	log.With("err", err).Error(msg)
}

func Errorw(msg string, err error, context ...interface{}) {
	context = append(context, "err")
	context = append(context, err)
	log.With(context...).Error(msg)
}

func Info(msg string) {
	log.Info(msg)
}

func Infow(msg string, context ...interface{}) {
	log.With(context...).Info(msg)
}

func Debug(msg string) {
	log.Debug(msg)
}

func Debugw(msg string, context ...interface{}) {
	log.With(context...).Debug(msg)
}
