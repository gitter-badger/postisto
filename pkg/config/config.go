package config

import (
	imapClient "github.com/emersion/go-imap/client"
	"github.com/goccy/go-yaml"
	"github.com/imdario/mergo"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Accounts map[string]*Account `yaml:"accounts"`

	Settings struct {
		//logging
		UseGPGAgent bool `yaml:"gpg_use_agent"`
	} `yaml:"settings"`
}

type Account struct {
	Connection AccountConnection `yaml:"connection"`
	Filters    map[string]Filter `yaml:"filters"`
}

type AccountConnection struct {
	Enabled         bool               `yaml:"enabled"`
	Server          string             `yaml:"server"`
	Port            int                `yaml:"port"`
	Username        string             `yaml:"username"`
	Password        string             `yaml:"password"`
	InputMailbox    *InputMailbox      `yaml:"input"`
	FallbackMailbox string             `yaml:"fallback_mailbox"`
	IMAPS           bool               `yaml:"imaps"`
	Starttls        *bool              `yaml:"starttls"`
	TLSVerify       *bool              `yaml:"tlsverify"`
	TLSCACertFile   string             `yaml:"cacertfile"`
	Client          *imapClient.Client //TODO custom type?
	Debug           bool               `yaml:"debug"` //TODO => use with log setting/level!
}

type Filter struct {
	Commands Commands `yaml:"commands,flow"`
	//Rules    []Rule    `yaml:"rules"`
}
type Commands map[string]interface{}
type Rule []map[string]interface{}

type InputMailbox struct {
	Mailbox      string   `yaml:"mailbox"`
	WithoutFlags []string `yaml:"without_flags,flow"`
}

func New() Config {
	return Config{}
}

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

func (cfg Config) validate() (Config, error) {

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

		if acc.Connection.InputMailbox == nil {
			acc.Connection.InputMailbox = &InputMailbox{Mailbox: "INBOX", WithoutFlags: []string{"\\Seen", "\\Flagged"}}
		}
	}

	// Filters

	// Settings

	return cfg, nil
}
