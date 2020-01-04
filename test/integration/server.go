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

func NewAccount(t *testing.T, port int, starttls bool, imaps bool, tlsverify bool, cacertfile *string) *config.Account {

	require := require.New(t)
	r1 := rand.New(rand.NewSource(time.Now().UnixNano()))

	if cacertfile == nil {
		defaultcacert := "../../test/data/certs/ca.pem"
		cacertfile = &defaultcacert
	}

	acc := config.Account{
		Connection: config.AccountConnection{
			Enabled:  true,
			Server:   "localhost",
			Port:     port,
			Username: fmt.Sprintf("test-%v@example.com", r1.Intn(10000)),
			Password: "test",
			InputMailbox: &config.InputMailbox{
				Name:           "INBOX",
				SearchCriteria: "UNSEEN UNFLAGGED",
			},
			IMAPS:         imaps,
			Starttls:      &starttls,
			TLSVerify:     &tlsverify,
			TLSCACertFile: *cacertfile,
		},
	}

	redisClient, err := NewRedisClient()
	require.Nil(err)

	err = NewIMAPUser(&acc, redisClient)
	require.Nil(err)

	return &acc
}

func NewIMAPUser(acc *config.Account, redisClient *redis.Client) error {
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

func NewRedisClient() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	_, err := redisClient.Ping().Result()

	return redisClient, err
}
