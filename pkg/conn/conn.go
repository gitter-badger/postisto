package conn

import (
	"crypto/tls"
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/emersion/go-imap/client"
)

func Connect(acc config.Account) (*client.Client, error) {
	var c *client.Client
	var err error

	tlsConfig := &tls.Config{
		ServerName:         acc.Connection.Server,
		InsecureSkipVerify: !*acc.Connection.TLSVerify,
		MinVersion:         tls.VersionTLS12,
	}

	if acc.Connection.IMAPS {
		c, err = client.DialTLS(fmt.Sprintf("%v:%v", acc.Connection.Server, acc.Connection.Port), tlsConfig)

	} else {
		c, err = client.Dial(fmt.Sprintf("%v:%v", acc.Connection.Server, acc.Connection.Port))

		if err != nil {
			return c, err
		}

		if *acc.Connection.Starttls {
			err = c.StartTLS(tlsConfig)
		}
	}

	if err != nil {
		return c, err
	}

	err = c.Login(acc.Connection.Username, acc.Connection.Password)

	return c, err
}

func DisconnectAll(conns map[string]*client.Client) error {
	var err error
	for _, c := range conns {
		err = c.Logout() //TODO that's evil. we're overring err all the time
	}
	return err
}
