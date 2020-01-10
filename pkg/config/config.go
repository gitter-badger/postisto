package config

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/filter"
	"github.com/arnisoph/postisto/pkg/log"
	"github.com/arnisoph/postisto/pkg/server"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Accounts map[string]Account                  `yaml:"accounts"`
	Filters  map[string]map[string]filter.Filter `yaml:"filters"`
	Settings struct {
		LogConfig log.Config `yaml:"logging"`
		//UseGPGAgent bool       `yaml:"gpg_use_agent"`
	} `yaml:"settings"`
}

type Account struct {
	Connection      server.Connection   `yaml:"connection"`
	InputMailbox    *InputMailboxConfig `yaml:"input"`
	FallbackMailbox *string             `yaml:"fallback_mailbox"`
}

type InputMailboxConfig struct {
	Mailbox      string   `yaml:"mailbox"`
	WithoutFlags []string `yaml:"without_flags,flow"`
}

func NewConfig() *Config {
	return new(Config)
}

func NewConfigWithDefaults() (*Config, error) {
	cfg := NewConfig()
	return cfg.validate()
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
		fileCfg := new(Config)
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

	return cfg.validate()
}

func (cfg Config) validate() (*Config, error) {
	valCfg := Config{
		Accounts: map[string]Account{},
		Filters:  map[string]map[string]filter.Filter{},
	}

	// Accounts
	if len(cfg.Accounts) == 0 {
		log.Info("Warning: no accounts configured")
	}

	for accName, acc := range cfg.Accounts {
		newAcc := Account{
			Connection:      acc.Connection,
			InputMailbox:    acc.InputMailbox,
			FallbackMailbox: acc.FallbackMailbox,
		}
		// Connection
		if acc.Connection.Server == "" {
			return nil, fmt.Errorf("server not configured")
		}

		// Input
		if newAcc.InputMailbox == nil || newAcc.InputMailbox.Mailbox == "" {
			newAcc.InputMailbox = new(InputMailboxConfig)
			newAcc.InputMailbox.Mailbox = "INBOX"
			newAcc.InputMailbox.WithoutFlags = []string{"\\Seen", "\\Flagged"}
		}

		if newAcc.FallbackMailbox == nil {
			newAcc.InputMailbox.WithoutFlags = []string{"\\Seen", "\\Flagged"}

			newAcc.FallbackMailbox = new(string)
			*newAcc.FallbackMailbox = "INBOX"
		}

		if *newAcc.FallbackMailbox != newAcc.InputMailbox.Mailbox {
			newAcc.InputMailbox.WithoutFlags = []string{}
		}

		valCfg.Accounts[accName] = newAcc
	}

	// Filters
	valCfg.Filters = cfg.Filters

	// Settings
	valCfg.Settings = cfg.Settings
	if cfg.Settings.LogConfig.PreSetMode == "" && cfg.Settings.LogConfig.ZapConfig == nil {
		valCfg.Settings.LogConfig.PreSetMode = "prod"
	}

	if cfg.Settings.LogConfig.PreSetMode != "" && cfg.Settings.LogConfig.ZapConfig != nil {
		return nil, fmt.Errorf("log config validation error: either set mode or config")
	}

	return &valCfg, nil
}
