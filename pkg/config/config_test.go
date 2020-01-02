package config

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestGetConfig(t *testing.T) {
	assert := assert.New(t)

	// Load single file
	cfg, err := GetConfig("../../tests/configs/valid/accounts.yaml")
	assert.Nil(err)
	assert.Equal("imap.server.de", cfg.Accounts["test"].Connection.Server)

	// Load full config dir
	cfg, err = GetConfig("../../tests/configs/valid/")
	assert.Nil(err)
	assert.Equal("imap.server.de", cfg.Accounts["test"].Connection.Server)

	// Failed file/dir loading
	cfg, err = GetConfig("../../tests/configs/does-not-exist")
	assert.EqualError(err, "stat ../../tests/configs/does-not-exist: no such file or directory")

	cfg, err = GetConfig("../../tests/configs/invalid/zero-file.yaml")
	assert.EqualError(err, "String node doesn't MapNode")

	_, _ = os.Create("../../tests/configs/invalid-unreadable-file/unreadable-file.testfile.yaml")
	err = os.Chmod("../../tests/configs/invalid-unreadable-file/unreadable-file.testfile.yaml", 0000)
	assert.Nil(err)
	cfg, err = GetConfig("../../tests/configs/invalid-unreadable-file/")
	assert.EqualError(err, "open ../../tests/configs/invalid-unreadable-file/unreadable-file.testfile.yaml: permission denied")
}
