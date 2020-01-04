package conn

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestConnect(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	redisClient, err := integration.NewRedisClient()
	require.Nil(err)
	accs := map[string]*config.Account{
		"starttls": integration.NewAccount(10143, true, false, true),
		"starttls_wrongport": integration.NewAccount(42, true, false, true),
		"imaps": integration.NewAccount(10993, false, true, true),
		"imaps_wrongport": integration.NewAccount(42, false, true, true),
	}

	require.Nil(integration.NewIMAPUser(accs["starttls"], redisClient))
	require.Nil(integration.NewIMAPUser(accs["starttls_wrongport"], redisClient), )
	require.Nil(integration.NewIMAPUser(accs["imaps"], redisClient))
	require.Nil(integration.NewIMAPUser(accs["imaps_wrongport"], redisClient))

	require.Nil(Connect(accs["starttls"]))
	assert.Error(Connect(accs["starttls_wrongport"]))
	require.Nil(Connect(accs["imaps"]))
	assert.Error(Connect(accs["imaps_wrongport"]))

	defer func() {
		for _, err := range DisconnectAll(accs) { //TODO verify whether accs actually contians all accs
			assert.Nil(err)
		}
	}()
}

func TestDisconnectAll(t *testing.T) {
	//assert := assert.New(t)
}
