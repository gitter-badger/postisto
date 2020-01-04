package conn

import (
	"crypto/tls"
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/emersion/go-imap/client"
)

func Connect(acc *config.Account) error {
	var err error

	tlsConfig := &tls.Config{
		ServerName:         acc.Connection.Server,
		InsecureSkipVerify: !*acc.Connection.TLSVerify, //TODO that's incredibely dangerous! do validation ourself here?
		MinVersion:         tls.VersionTLS12,
	}

	if acc.Connection.IMAPS {
		acc.Connection.Client, err = client.DialTLS(fmt.Sprintf("%v:%v", acc.Connection.Server, acc.Connection.Port), tlsConfig)

	} else {
		acc.Connection.Client, err = client.Dial(fmt.Sprintf("%v:%v", acc.Connection.Server, acc.Connection.Port))

		if err != nil {
			return err
		}

		if *acc.Connection.Starttls {
			err = acc.Connection.Client.StartTLS(tlsConfig)
		}
	}

	if err != nil {
		return err
	}

	err = acc.Connection.Client.Login(acc.Connection.Username, acc.Connection.Password)

	return err
}

func DisconnectAll(accs map[string]*config.Account) map[string]error {
	err := map[string]error{}

	for name, acc := range accs {
		if acc.Connection.Client == nil {
			// no connection
			continue
		}
		err[name] = acc.Connection.Client.Logout()
	}
	return err
}
