package config

import (
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGetConfig(t *testing.T) {
	require := require.New(t)

	cfg := NewConfig()
	var err error

	// ACTUAL TESTS BELOW

	// Load single file
	require.FileExists("../../test/data/configs/valid/accounts.yaml")
	cfg, err = cfg.Load("../../test/data/configs/valid/accounts.yaml")
	require.Nil(err)
	require.Equal("imap.server.de", cfg.Accounts["test"].Connection.Server)

	// Load full config dir
	require.DirExists("../../test/data/configs/valid/")
	cfg, err = cfg.Load("../../test/data/configs/valid/")
	require.Nil(err)
	//fmt.Println(yaml.FormatError(err, true, true))
	require.Equal("imap.server.de", cfg.Accounts["test"].Connection.Server)

	// Failed file/dir loading
	cfg, err = cfg.Load("../../test/data/configs/does-not-exist")
	require.EqualError(err, "stat ../../test/data/configs/does-not-exist: no such file or directory")

	// Reading broken file
	cfg, err = cfg.Load("../../test/data/configs/invalid/zero-file.yaml")
	require.EqualError(err, "String node doesn't MapNode")

	// Reading unaccessible file
	_, _ = os.Create("../../test/data/configs/invalid-unreadable-file/unreadable-file.testfile.yaml")
	require.Nil(os.Chmod("../../test/data/configs/invalid-unreadable-file/unreadable-file.testfile.yaml", 0000))
	cfg, err = cfg.Load("../../test/data/configs/invalid-unreadable-file/")
	require.EqualError(err, "open ../../test/data/configs/invalid-unreadable-file/unreadable-file.testfile.yaml: permission denied")
}

func TestFilterSet_Names(t *testing.T) {
	require := require.New(t)

	cfg := NewConfig()
	var err error

	// ACTUAL TESTS BELOW

	// load test data
	cfg, err = cfg.Load("../../test/data/configs/valid/local_imap_server/shops.yaml")
	require.Nil(err)

	// test our funcy func
	require.ElementsMatch(cfg.Accounts["local_imap_server"].FilterSet.Names(), []string{"shops"})
}
