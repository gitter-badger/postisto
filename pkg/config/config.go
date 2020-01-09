package config

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/log"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
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

func NewConfigFromFile(configPath string) (*Config, error) {
	cfg := NewConfig()
	var configFiles []string

	stat, err := os.Stat(configPath)
	if err != nil {
		log.Errorw("Failed to check path", err, "configPath", configPath)
		return nil, err
	}

	if stat.IsDir() {
		err := filepath.Walk(configPath,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					log.Errorw("Failed to load path", err, "path", path)
					return err
				}

				if stat, err := os.Stat(path); err != nil {
					log.Errorw("Failed to load path", err, "path", path)
					return err
				} else if !stat.IsDir() && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
					configFiles = append(configFiles, path)
				}

				return nil
			})

		if err != nil {
			log.Errorw("Failed to parse dir", err, "configPath", configPath)
			return nil, err
		}
	} else {
		configFiles = append(configFiles, configPath)
	}

	for _, file := range configFiles {
		fileCfg := Config{}
		yamlFile, err := ioutil.ReadFile(file)

		if err != nil {
			log.Errorw("Failed to read file", err, "file", file)
			return nil, err
		}

		err = yaml.Unmarshal(yamlFile, &fileCfg)

		if err != nil {
			log.Errorw("Failed to parse YAML file", err, "file", file)
			return nil, err
		}

		if err := mergo.Merge(cfg, fileCfg, mergo.WithOverride); err != nil {
			log.Errorw("Failed to merge YAML file", err, "file", file)
			return nil, err
		}
	}

	cfg.setDefaults()
	if err = cfg.validate(); err != nil {
		log.Errorw("Failed to validate YAML", err, "yaml", cfg)
		return nil, err
	}

	return cfg, nil
}

func (cfg *Config) validate() error {

	// Accounts

	// Filters

	// Settings
	if cfg.Settings.LogConfig.PreSetMode != "" && cfg.Settings.LogConfig.ZapConfig != nil {
		return fmt.Errorf("log config validation error: either set mode or config")
	}

	return nil
}

func (cfg *Config) setDefaults() {
	// Accounts
	for _, acc := range cfg.Accounts {
		if acc == nil {
			acc = new(Account)
		}

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

		if acc.InputMailbox == nil || acc.InputMailbox.Mailbox == "" {
			acc.InputMailbox = &InputMailbox{Mailbox: "INBOX", WithoutFlags: []string{"\\Seen", "\\Flagged"}}
		}

		if acc.FallbackMailbox == nil {
			fallback := "INBOX"
			acc.FallbackMailbox = &fallback
		}

		if acc.InputMailbox.Mailbox == *acc.FallbackMailbox && len(acc.InputMailbox.WithoutFlags) == 0 {
			acc.InputMailbox.WithoutFlags = []string{"\\Seen", "\\Flagged"}
		}
	}

	// Filters

	// Settings
	// && len(cfg.Settings.LogConfig.ZapConfig.OutputPaths) == 0
	if cfg.Settings.LogConfig.PreSetMode == "" && cfg.Settings.LogConfig.ZapConfig == nil {
		cfg.Settings.LogConfig.PreSetMode = "prod"
	}
}
