package config

import (
	"github.com/emersion/go-imap"
	imapClient "github.com/emersion/go-imap/client"
	"go.uber.org/zap"
)

type Config struct {
	Accounts map[string]*Account `yaml:"accounts"`

	Settings struct {
		LogConfig LogConfig `yaml:"logging"`
		UseGPGAgent bool `yaml:"gpg_use_agent"`
	} `yaml:"settings"`
}

type LogConfig struct {
	PreSetMode string      `yaml:"mode"`
	ZapConfig     *zap.Config `yaml:"dangerousAdvancedZapConfig"`
}

type Account struct {
	Connection      ConnectionConfig `yaml:"connection"`
	InputMailbox    *InputMailbox    `yaml:"input"`
	FallbackMailbox *string          `yaml:"fallback_mailbox"`
	FilterSet       FilterSet        `yaml:"filters"`
	Debug           bool             `yaml:"debug"` //TODO => use with log setting/level!
}

type ConnectionConfig struct {
	Enabled       bool               `yaml:"enabled"`
	Server        string             `yaml:"server"`
	Port          int                `yaml:"port"`
	Username      string             `yaml:"username"`
	Password      string             `yaml:"password"`
	IMAPS         bool               `yaml:"imaps"`
	Starttls      *bool              `yaml:"starttls"`
	TLSVerify     *bool              `yaml:"tlsverify"`
	TLSCACertFile string             `yaml:"cacertfile"`
	Client        *imapClient.Client //TODO custom type?
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

type Mail struct {
	RawMail *imap.Message
	Headers MailHeaders
}
type MailHeaders map[string]interface{}

func NewMail(rawMail *imap.Message, headers MailHeaders) Mail {
	return Mail{RawMail: rawMail, Headers: headers}
}

//func (msg Mail) GetHeaders(c *imapClient.Client) MailHeaders {
//	if len(msg.headers) == 0 {
//		mail.ParseMailHeaders(c, []config.Mailmsg.RawMail)
//	}
//}
