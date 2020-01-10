package main

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/filter"
	"github.com/arnisoph/postisto/pkg/log"
	"github.com/emersion/go-imap"
	"github.com/urfave/cli/v2"
	goLog "log"
	"os"
	"time"
)

func main() {
	var configPath string
	var logLevel string
	var logJSON bool

	app := &cli.App{
		Name:  "poŝtisto",
		Usage: "foo ce",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "config file or directory path",
				Value:       "config/",
				EnvVars:     []string{"CONFIG_PATH"},
				Destination: &configPath,
			},
			&cli.StringFlag{
				Name:        "log-level",
				Aliases:     []string{"l"},
				Usage:       "log level e.g. trace, debug, info or error (WARNING: trace exposes account credentials and mor sensitive data)",
				Value:       "info",
				EnvVars:     []string{"LOG_LEVEL"},
				Destination: &logLevel,
			},
			&cli.BoolFlag{
				Name:        "log-json",
				Aliases:     []string{"j"},
				Usage:       "format log output as JSON",
				Value:       false,
				EnvVars:     []string{"LOG_JSON"},
				Destination: &logJSON,
				HasBeenSet:  false,
			},
		},
		Action: func(c *cli.Context) error {
			return startApp(c, configPath, logLevel, logJSON)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		goLog.Fatalln("Failed to start app:", err)
	}
}

func startApp(c *cli.Context, configPath string, logLevel string, logJSON bool) error {

	if err := log.InitWithConfig(logLevel, logJSON); err != nil {
		return err
	}

	var cfg *config.Config
	var err error

	if cfg, err = config.NewConfigFromFile(configPath); err != nil {
		return err
	}

	if len(cfg.Accounts) == 0 {
		return fmt.Errorf("no (enabled) account configuration found. nothing to do")
	}

	if len(cfg.Filters) == 0 {
		return fmt.Errorf("no filter configuration found. nothing to do")
	}

	for {
		for name, acc := range cfg.Accounts {
			filters, ok := cfg.Filters[name]
			if !ok {
				return fmt.Errorf("no filter configuration found for account %v. nothing to do", name)
			}

			if err := acc.Connection.Connect(); err != nil {
				return err
			}

			if err := filter.EvaluateFilterSetsOnMsgs(&acc.Connection, *acc.InputMailbox, []string{imap.SeenFlag, imap.FlaggedFlag}, *acc.FallbackMailbox, filters); err != nil {
				return fmt.Errorf("failed to run filter engine: %v", err)
			}
		}

		time.Sleep(time.Second * 10)
	}
}
