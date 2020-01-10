package log_test

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/log"

	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitWithConfig(t *testing.T) {
	require := require.New(t)

	// ACTUAL TESTS BELOW

	// Prepare some test log events
	testLogging := func() {
		log.Debug("Testing a debug log event.")
		log.Debugw("Testing a debug log event.", "with", "fields", "numeric work too", 42, "or even maps", map[string]string{"foo": "bar"}, "why not also try slices", []int{1, 3, 3, 7})
		log.Info("Testing an info log event.")
		log.Infow("Testing an info log event.", "with", "fields", "numeric work too", 42, "or even maps", map[string]string{"foo": "bar"}, "why not also try slices", []int{1, 3, 3, 7})
		log.Error("Testing an error log event.", fmt.Errorf("test error"))
		log.Errorw("Testing an error log event.", fmt.Errorf("test error"), "with", "fields", "numeric work too", 42, "or even maps", map[string]string{"foo": "bar"}, "why not also try slices", []int{1, 3, 3, 7})
	}

	fmt.Println("Log with full user-defined log config (zap.Config)")
	cfg, err := config.NewConfigFromFile("../../test/data/configs/valid/test/CustomLogConfig.yaml")
	require.NoError(err)
	require.NotNil(cfg)

	require.NoError(log.InitWithConfig(cfg.Settings.LogConfig))
	testLogging()

	fmt.Println("Log with program default (preset mode debug)")
	testLogging()

	fmt.Println("Log with preset mode dev")
	cfg, err = config.NewConfigWithDefaults()
	require.NoError(err)
	require.NotNil(cfg)
	cfg.Settings.LogConfig.PreSetMode = "dev"
	require.NoError(log.InitWithConfig(cfg.Settings.LogConfig))
	testLogging()

	fmt.Println("Log with user default (preset mode prod)")
	cfg, err = config.NewConfigWithDefaults()
	require.NoError(err)
	require.NotNil(cfg)
	require.NoError(log.InitWithConfig(cfg.Settings.LogConfig))
	testLogging()
}
