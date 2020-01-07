package conn

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	imapClient "github.com/emersion/go-imap/client"

	"io/ioutil"
)

func Connect(connConfig config.ConnectionConfig) (*imapClient.Client, error) {
	var c *imapClient.Client
	var err error

	certPool := x509.NewCertPool()
	if connConfig.TLSCACertFile != "" {
		pemBytes, err := ioutil.ReadFile(connConfig.TLSCACertFile)
		if err != nil {
			return nil, err
		}

		certPool.AppendCertsFromPEM(pemBytes)

	} else {
		certPool = nil
	}

	tlsConfig := &tls.Config{
		ServerName:         connConfig.Server,
		InsecureSkipVerify: !*connConfig.TLSVerify, //TODO that's incredibely dangerous! do validation ourself here?
		MinVersion:         tls.VersionTLS12,
		RootCAs:            certPool,
	}

	if connConfig.IMAPS {
		c, err = imapClient.DialTLS(fmt.Sprintf("%v:%v", connConfig.Server, connConfig.Port), tlsConfig)

	} else {
		c, err = imapClient.Dial(fmt.Sprintf("%v:%v", connConfig.Server, connConfig.Port))

		if err != nil {
			return nil, err
		}

		if *connConfig.Starttls {
			err = c.StartTLS(tlsConfig)
		}
	}

	if err != nil {
		return nil, err
	}

	if true { //TODO
		//c.SetDebug(os.Stderr)
	}

	if err = c.Login(connConfig.Username, connConfig.Password); err != nil {
		return nil, err
	}

	return c, err
}

func Disconnect(c *imapClient.Client) error {

	if c == nil {
		// no connection
		return nil
	}
	return c.Logout()
}
