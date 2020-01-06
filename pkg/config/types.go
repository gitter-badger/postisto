package config

import (
	"github.com/emersion/go-imap"
	imapClient "github.com/emersion/go-imap/client"
)

type Config struct {
	Accounts map[string]*Account `yaml:"accounts"`

	Settings struct {
		//logging
		UseGPGAgent bool `yaml:"gpg_use_agent"`
	} `yaml:"settings"`
}

func NewConfig() Config {
	return Config{}
}

type Account struct {
	Connection ConnectionConfig `yaml:"connection"`
	FilterSet  FilterSet        `yaml:"filters"`
}

type ConnectionConfig struct {
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
	Commands Commands `yaml:"commands,flow"`
	RuleSet  RuleSet  `yaml:"rules"`
}
type Commands map[string]interface{}
type RuleSet []Rule
type Rule map[string][]map[string]interface{}

type InputMailbox struct {
	Mailbox      string   `yaml:"mailbox"`
	WithoutFlags []string `yaml:"without_flags,flow"`
}

type Mail struct {
	RawMail *imap.Message
	Headers MailHeaders
}
type MailHeaders map[string]string

func NewMail(rawMail *imap.Message) Mail {
	return Mail{RawMail: rawMail, Headers: MailHeaders{}}
}
