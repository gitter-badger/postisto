package conn

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/stretchr/testify/require"
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
			require.Nil(Disconnect(acc))
		}
	}()

	// ACTUAL TESTS BELOW

	// connect to IMAP server
	require.Nil(Connect(accs["starttls"]))
	require.Error(Connect(accs["starttls_wrongport"]))
	require.Nil(Connect(accs["imaps"]))
	require.Error(Connect(accs["imaps_wrongport"]))
	//require.EqualError(Connect(accs["nocacert"]), "x509: certificate signed by unknown authority")
	require.EqualError(Connect(accs["badcacert"]), "x509: certificate signed by unknown authority")
	require.EqualError(Connect(accs["badcacertpath"]), "open ca-doesnotexist.pem: no such file or directory")
}
