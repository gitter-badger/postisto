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
		"main": integration.NewAccount(10143, true, false, true),
		"main_imaps": integration.NewAccount(10993, false, true, true),
	}

	require.Nil(integration.NewIMAPUser(accs["main"], redisClient))
	require.Nil(integration.NewIMAPUser(accs["main_imaps"], redisClient))

	require.Nil(Connect(accs["main"]))
	require.Nil(Connect(accs["main_imaps"]))

	defer func() {
		for _, err := range DisconnectAll(accs) { //TODO verify whether accs actually contians all accs
			assert.Nil(err)
		}
	}()
}

func TestDisconnectAll(t *testing.T) {
	//assert := assert.New(t)
}
