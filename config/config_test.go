package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestGetConfig(t *testing.T) {
	assert := assert.New(t)

	// Load single file
	cfg, err := GetConfig("../tests/configs/valid/accounts.yaml")
	assert.Nil(err)
	assert.Equal("imap.server.de", cfg.Accounts["test"].Server)

	// Load full config dir
	cfg, err = GetConfig("../tests/configs/valid/")
	assert.Nil(err)
	assert.Equal("imap.server.de", cfg.Accounts["test"].Server)
}
