package conn

import (
	"github.com/arnisoph/postisto/pkg/config"
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

	accs := map[string]*config.Account{
		"starttls":           integration.NewAccount(t, 10143, true, false, true, nil),
		"starttls_wrongport": integration.NewAccount(t, 42, true, false, true, nil),
		"imaps":              integration.NewAccount(t, 10993, false, true, true, nil),
		"imaps_wrongport":    integration.NewAccount(t, 42, false, true, true, nil),
		"nocacert":           integration.NewAccount(t, 10143, true, false, true, &nocacert),
		"badcacert":          integration.NewAccount(t, 10143, true, false, true, &badcacert),
		"badcacertpath":      integration.NewAccount(t, 10143, true, false, true, &badcacertpath),
	}

	defer func() {
		for _, acc := range accs {
			require.Nil(Disconnect(acc.Connection.Client))
		}
	}()

	// ACTUAL TESTS BELOW

	var err error

	// connect to IMAP server
	accs["starttls"].Connection.Client, err = Connect(accs["starttls"].Connection)
	require.Nil(err)
	accs["starttls_wrongport"].Connection.Client, err = Connect(accs["starttls_wrongport"].Connection)
	require.Error(err)
	accs["imaps"].Connection.Client, err = Connect(accs["imaps"].Connection)
	require.Nil(err)
	accs["imaps_wrongport"].Connection.Client, err = Connect(accs["imaps_wrongport"].Connection)
	require.Error(err)
	if os.Getenv("USER") != "ab" {
		_, err = Connect(accs["nocacert"].Connection)
		require.EqualError(err, "x509: certificate signed by unknown authority")
	}
	accs["badcacert"].Connection.Client, err = Connect(accs["badcacert"].Connection)
	require.EqualError(err, "x509: certificate signed by unknown authority")
	accs["badcacertpath"].Connection.Client, err = Connect(accs["badcacertpath"].Connection)
	require.EqualError(err, "open ca-doesnotexist.pem: no such file or directory")
}
