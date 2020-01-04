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
	accs := map[string]*config.Account{"main": integration.NewAccount()}
	err = integration.NewIMAPUser(accs["main"], redisClient)
	require.Nil(err)

	err = Connect(accs["main"])
	require.Nil(err)
	defer func() {
		for _, err := range DisconnectAll(accs) {
			assert.Nil(err)
		}
	}()
}

func TestDisconnectAll(t *testing.T) {
	//assert := assert.New(t)
}
