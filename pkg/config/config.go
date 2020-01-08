package config

import (
	"fmt"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func NewConfig() *Config {
	return &Config{}
}

func NewConfigWithDefaults() *Config {
	cfg := NewConfig()
	cfg.setDefaults()
	return cfg
}

func  NewConfigFromFile(configPath string) (*Config, error) {
	cfg := NewConfig()
	var configFiles []string

	stat, err := os.Stat(configPath)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		err := filepath.Walk(configPath,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if stat, _ := os.Stat(path); !stat.IsDir() && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
					configFiles = append(configFiles, path)
				}
				return err
			})
		if err != nil {
			return nil, err
		}
	} else {
		configFiles = append(configFiles, configPath)
	}

	for _, file := range configFiles {
		fileCfg := Config{}
		yamlFile, err := ioutil.ReadFile(file)

		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(yamlFile, &fileCfg)

		if err != nil {
			return nil, err
		}

		if err := mergo.Merge(cfg, fileCfg, mergo.WithOverride); err != nil {
			return nil, err
		}
	}

	cfg.setDefaults()
	if err = cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) validate() error {

	// Accounts

	// Filters

	// Settings
	if cfg.Settings.LogConfig.PreSetMode != "" && cfg.Settings.LogConfig.ZapConfig != nil {
		return fmt.Errorf("log config validation error: either set mode or config") //TODO
	}

	return nil
}

func (cfg *Config) setDefaults() {
	// Accounts
	for _, acc := range cfg.Accounts {
		// When not using IMAPS, enable STARTTLS by default
		if !acc.Connection.IMAPS && acc.Connection.Starttls == nil {
			var b bool
			acc.Connection.Starttls = &b
			*acc.Connection.Starttls = true
		}

		if acc.Connection.TLSVerify == nil {
			var b bool
			acc.Connection.TLSVerify = &b
			*acc.Connection.TLSVerify = true
		}

		if acc.InputMailbox == nil {
			acc.InputMailbox = &InputMailbox{Mailbox: "INBOX", WithoutFlags: []string{"\\Seen", "\\Flagged"}}
		}

		if acc.FallbackMailbox == nil {
			fallback := "INBOX"
			acc.FallbackMailbox = &fallback
		}
	}

	// Filters

	// Settings
	if cfg.Settings.LogConfig.PreSetMode == "" && len(cfg.Settings.LogConfig.ZapConfig.OutputPaths) == 0 {
		cfg.Settings.LogConfig.PreSetMode = "prod"
	}
}