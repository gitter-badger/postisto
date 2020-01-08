package integration

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/redis.v4"
	"math/rand"
	"testing"
	"time"
)

const MaxTestMailCount = 17

func NewAccount(t *testing.T, username string, password string, port int, starttls bool, imaps bool, tlsverify bool, cacertfile *string) *config.Account {

	require := require.New(t)

	if cacertfile == nil {
		defaultcacert := "../../test/data/certs/ca.pem"
		cacertfile = &defaultcacert
	}

	acc := config.Account{
		InputMailbox: &config.InputMailbox{
			Mailbox:      "INBOX",
			WithoutFlags: []string{"\\Seen", "\\Flagged"},
		},
		Connection: config.ConnectionConfig{
			Enabled:  true,
			Server:   "localhost",
			Port:     port,
			Username: NewUsername(username),
			Password: password,

			IMAPS:         imaps,
			Starttls:      &starttls,
			TLSVerify:     &tlsverify,
			TLSCACertFile: *cacertfile,
			DebugIMAP:     false,
		},
	}

	redisClient, err := newRedisClient()
	require.Nil(err)

	err = newIMAPUser(&acc, redisClient)
	require.Nil(err)

	return &acc
}

func NewStandardAccount(t *testing.T) *config.Account {
	return NewAccount(t, "", "test", 10143, true, false, true, nil)
}

func NewUsername(u string) string {
	if u != "" {
		return u
	}

	r1 := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("test-%v@example.com", r1.Intn(1000000)) //TODO that's not random enough
}

func newIMAPUser(acc *config.Account, redisClient *redis.Client) error {
	dbs := [2]string{"userdb", "passdb"}
	for _, db := range dbs {
		key := fmt.Sprintf("dovecot/%v/%v", db, acc.Connection.Username)
		value := fmt.Sprintf(`{"uid":"65534","gid":"65534","home":"/tmp/%[1]v","username":"%[1]v","password":"%[2]v"}`, acc.Connection.Username, acc.Connection.Password)

		if err := redisClient.Set(key, value, 0).Err(); err != nil {
			return err
		}
	}

	return nil
}

func newRedisClient() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping().Result()

	return redisClient, err
}
