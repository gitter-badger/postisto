package config

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGetConfig(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	cfg := New()
	var err error

	// Load single file
	require.FileExists("../../test/data/configs/valid/accounts.yaml")
	cfg, err = cfg.Load("../../test/data/configs/valid/accounts.yaml")
	require.Nil(err)
	assert.Equal("imap.server.de", cfg.Accounts["test"].Connection.Server)

	// Load full config dir
	require.DirExists("../../test/data/configs/valid/")
	cfg, err = cfg.Load("../../test/data/configs/valid/")
	//fmt.Println(yaml.FormatError(err, true, true))
	require.Nil(err)
	assert.Equal("imap.server.de", cfg.Accounts["test"].Connection.Server)

	// Failed file/dir loading
	cfg, err = cfg.Load("../../test/data/configs/does-not-exist")
	assert.EqualError(err, "stat ../../test/data/configs/does-not-exist: no such file or directory")

	cfg, err = cfg.Load("../../test/data/configs/invalid/zero-file.yaml")
	assert.EqualError(err, "String node doesn't MapNode")

	_, _ = os.Create("../../test/data/configs/invalid-unreadable-file/unreadable-file.testfile.yaml")
	require.Nil(os.Chmod("../../test/data/configs/invalid-unreadable-file/unreadable-file.testfile.yaml", 0000))
	cfg, err = cfg.Load("../../test/data/configs/invalid-unreadable-file/")
	assert.EqualError(err, "open ../../test/data/configs/invalid-unreadable-file/unreadable-file.testfile.yaml: permission denied")
}
