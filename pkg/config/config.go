package config

import (
	"github.com/goccy/go-yaml"
	"github.com/imdario/mergo"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func (cfg Config) Load(configPath string) (Config, error) {

	var configFiles []string

	stat, err := os.Stat(configPath)
	if err != nil {
		return cfg, err
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
				return nil
			})
		if err != nil {
			return cfg, err
		}
	} else {
		configFiles = append(configFiles, configPath)
	}

	for _, file := range configFiles {
		fileCfg := Config{}
		yamlFile, err := ioutil.ReadFile(file)

		if err != nil {
			return cfg, err
		}

		err = yaml.Unmarshal(yamlFile, &fileCfg)

		if err != nil {
			return cfg, err
		}

		if err := mergo.Merge(&cfg, fileCfg, mergo.WithOverride); err != nil {
			return cfg, err
		}
	}

	return cfg.validate()
}

func (cfg Config) validate() (Config, error) { //TODO

	// Accounts
	for i := range cfg.Accounts {
		acc := cfg.Accounts[i]

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
	}

	// Filters

	// Settings

	return cfg, nil
}
