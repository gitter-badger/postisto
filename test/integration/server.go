package integration

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"gopkg.in/redis.v4"
	"math/rand"
	"time"
)

func NewAccount() *config.Account {

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	starttls := true
	tlsverify := true

	acc := config.Account{
		Connection: config.AccountConnection{
			Enabled:  true,
			Server:   "localhost",
			Port:     10143,
			Username: fmt.Sprintf("test-%v@example.com", r1.Intn(10000)),
			Password: "test",
			InputMailbox: &config.InputMailbox{
				Name:           "INBOX",
				SearchCriteria: "UNSEEN UNFLAGGED",
			},
			IMAPS:         false,
			Starttls:      &starttls,
			TLSVerify:     &tlsverify,
			TLSCACertFile: "../../test/data/ca.pem",
		},
	}

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
