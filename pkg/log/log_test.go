package log

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestInitWithConfig(t *testing.T) {
	require := require.New(t)

	// ACTUAL TESTS BELOW

	// Prepare some test log events
	testLogging := func() {
		Debug("Testing a debug log event.")
		Debugw("Testing a debug log event.", "with", "fields", "numeric work too", 42, "or even maps", map[string]string{"foo": "bar"}, "why not also try slices", []int{1, 3, 3, 7})
		Info("Testing an info log event.")
		Infow("Testing an info log event.", "with", "fields", "numeric work too", 42, "or even maps", map[string]string{"foo": "bar"}, "why not also try slices", []int{1, 3, 3, 7})
		Error("Testing an error log event.")
		Errorw("Testing an error log event.", "with", "fields", "numeric work too", 42, "or even maps", map[string]string{"foo": "bar"}, "why not also try slices", []int{1, 3, 3, 7})
	}

	fmt.Println("Log with program default (preset mode debug)")
	testLogging()

	fmt.Println("Log with preset mode dev")
	cfg := config.NewConfigWithDefaults()
	require.NotNil(cfg)
	cfg.Settings.LogConfig.PreSetMode = "dev"
	require.NoError(InitWithConfig(cfg.Settings.LogConfig))
	testLogging()

	fmt.Println("Log with user default (preset mode prod)")
	cfg = config.NewConfigWithDefaults()
	require.NotNil(cfg)
	require.NoError(InitWithConfig(cfg.Settings.LogConfig))
	testLogging()
}
