package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGetConfig(t *testing.T) {

	require := require.New(t)

	cfg := New()
	var err error

	// Load single file
	require.FileExists("../../test/data/configs/valid/accounts.yaml")
	cfg, err = cfg.Load("../../test/data/configs/valid/accounts.yaml")
	require.Nil(err)
	require.Equal("imap.server.de", cfg.Accounts["test"].Connection.Server)

	// Load full config dir
	require.DirExists("../../test/data/configs/valid/")
	cfg, err = cfg.Load("../../test/data/configs/valid/")
	//fmt.Println(yaml.FormatError(err, true, true))
	require.Nil(err)
	require.Equal("imap.server.de", cfg.Accounts["test"].Connection.Server)

	// Failed file/dir loading
	cfg, err = cfg.Load("../../test/data/configs/does-not-exist")
	require.EqualError(err, "stat ../../test/data/configs/does-not-exist: no such file or directory")

	cfg, err = cfg.Load("../../test/data/configs/invalid/zero-file.yaml")
	require.EqualError(err, "String node doesn't MapNode")

	_, _ = os.Create("../../test/data/configs/invalid-unreadable-file/unreadable-file.testfile.yaml")
	require.Nil(os.Chmod("../../test/data/configs/invalid-unreadable-file/unreadable-file.testfile.yaml", 0000))
	cfg, err = cfg.Load("../../test/data/configs/invalid-unreadable-file/")
	require.EqualError(err, "open ../../test/data/configs/invalid-unreadable-file/unreadable-file.testfile.yaml: permission denied")
}
