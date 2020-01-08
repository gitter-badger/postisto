package config

import "github.com/arnisoph/postisto/pkg/log"

type Config struct {
	Accounts map[string]*Account `yaml:"accounts"`

	Settings struct {
		LogConfig   log.Config `yaml:"logging"`
		UseGPGAgent bool       `yaml:"gpg_use_agent"`
	} `yaml:"settings"`
}

type Account struct {
	Connection      ConnectionConfig `yaml:"connection"`
	InputMailbox    *InputMailbox    `yaml:"input"`
	FallbackMailbox *string          `yaml:"fallback_mailbox"`
	FilterSet       FilterSet        `yaml:"filters"`
}

type ConnectionConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Server        string `yaml:"server"`
	Port          int    `yaml:"port"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
	IMAPS         bool   `yaml:"imaps"`
	Starttls      *bool  `yaml:"starttls"`
	TLSVerify     *bool  `yaml:"tlsverify"`
	TLSCACertFile string `yaml:"cacertfile"`
	DebugIMAP     bool   `yaml:"debug"`
}

type FilterSet map[string]Filter

func (filterSet FilterSet) Names() []string {
	keys := make([]string, len(filterSet))

	var i uint64
	for key, _ := range filterSet {
		keys[i] = key
		i++
	}

	return keys
}

type Filter struct {
	Commands FilterOps `yaml:"commands,flow"`
	RuleSet  RuleSet   `yaml:"rules"`
}
type FilterOps map[string]interface{}
type RuleSet []Rule
type Rule map[string][]map[string]interface{}

type InputMailbox struct {
	Mailbox      string   `yaml:"mailbox"`
	WithoutFlags []string `yaml:"without_flags,flow"`
}
