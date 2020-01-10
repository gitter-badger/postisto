package config

import (
	"github.com/arnisoph/postisto/pkg/filter"
	"github.com/arnisoph/postisto/pkg/log"
	"github.com/arnisoph/postisto/pkg/server"
)

type Config struct {
	Accounts map[string]Account   `yaml:"accounts"`
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

/*
type ConnectionClient interface {
//	List() (map[string]imapUtil.MailboxInfo, error)
	Connect(ConnectionConfig) (*imapClient.Connection, error)
}

*/
