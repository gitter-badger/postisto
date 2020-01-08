package imap_test

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/imap"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestConnect(t *testing.T) {
	require := require.New(t)

	nocacert := ""
	badcacert := "../../test/data/certs/bad-ca.pem"
	badcacertpath := "ca-doesnotexist.pem"

	accs := []struct {
		name   string
		acc    *config.Account
		client *imap.Client
	}{
		{name: "starttls", acc: integration.NewAccount(t, "", "", 10143, true, false, true, nil)},
		{name: "starttls_wrongport", acc: integration.NewAccount(t, "", "", 42, true, false, true, nil)},
		{name: "imaps", acc: integration.NewAccount(t, "", "", 10993, false, true, true, nil)},
		{name: "imaps_wrongport", acc: integration.NewAccount(t, "", "", 42, false, true, true, nil)},
		{name: "nocacert", acc: integration.NewAccount(t, "", "", 10143, true, false, true, &nocacert)},
		{name: "badcacert", acc: integration.NewAccount(t, "", "", 10143, true, false, true, &badcacert)},
		{name: "badcacertpath", acc: integration.NewAccount(t, "", "", 10143, true, false, true, &badcacertpath)},
	}

	defer func() {
		for _, acc := range accs {
			if acc.client != nil {
				require.Nil(acc.client.Disconnect())
			}
		}
	}()

	// ACTUAL TESTS BELOW
	var err error

	accs[0].client, err = imap.NewClient(accs[0].acc.Connection)
	require.Nil(err)

	accs[0].acc.Connection.Password = "wrongpass"
	accs[0].client, err = imap.NewClient(accs[0].acc.Connection)
	require.EqualError(err, "Authentication failed.")

	accs[1].client, err = imap.NewClient(accs[1].acc.Connection)
	require.Error(err)

	accs[2].client, err = imap.NewClient(accs[2].acc.Connection)
	require.Nil(err)

	accs[3].client, err = imap.NewClient(accs[3].acc.Connection)
	require.Error(err)

	if os.Getenv("USER") != "ab" {
		_, err = imap.NewClient(accs[4].acc.Connection)
		require.EqualError(err, "x509: certificate signed by unknown authority")
	}

	accs[5].client, err = imap.NewClient(accs[5].acc.Connection)
	require.EqualError(err, "x509: certificate signed by unknown authority")

	accs[6].client, err = imap.NewClient(accs[6].acc.Connection)
	require.EqualError(err, "open ca-doesnotexist.pem: no such file or directory")
}
