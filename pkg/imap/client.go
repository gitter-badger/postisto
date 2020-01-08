package imap

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/log"
	imapClient "github.com/emersion/go-imap/client"
	"io/ioutil"
	"os"
)

type Client struct {
	client *imapClient.Client
}

func NewClient(connConfig config.ConnectionConfig) (*Client, error) {
	var c *imapClient.Client
	var err error

	certPool := x509.NewCertPool()
	if connConfig.TLSCACertFile != "" {
		pemBytes, err := ioutil.ReadFile(connConfig.TLSCACertFile)
		if err != nil {
			log.Errorw("Failed to load CA cert file", err, "TLSCACertFile", connConfig.TLSCACertFile)
			return nil, err
		}

		certPool.AppendCertsFromPEM(pemBytes)

	} else {
		certPool = nil
	}

	tlsConfig := &tls.Config{
		ServerName:         connConfig.Server,
		InsecureSkipVerify: !*connConfig.TLSVerify,
		MinVersion:         tls.VersionTLS12,
		RootCAs:            certPool,
	}

	if connConfig.IMAPS {
		if c, err = imapClient.DialTLS(fmt.Sprintf("%v:%v", connConfig.Server, connConfig.Port), tlsConfig); err != nil {
			log.Errorw("Failed to connect to server", err, "server", connConfig.Server)
			return nil, err
		}
	} else {
		if c, err = imapClient.Dial(fmt.Sprintf("%v:%v", connConfig.Server, connConfig.Port)); err != nil {
			log.Errorw("Failed to connect to server", err, "server", connConfig.Server)
			return nil, err
		}

		if *connConfig.Starttls {
			if err = c.StartTLS(tlsConfig); err != nil {
				log.Errorw("Failed to initiate TLS session after connecting to server (STARTTLS)", err, "server", connConfig.Server)
				return nil, err
			}
		}
	}

	if connConfig.DebugIMAP {
		c.SetDebug(os.Stderr)
	}

	if err = c.Login(connConfig.Username, connConfig.Password); err != nil {
		log.Errorw("Failed to login to server", err, "server", connConfig.Server, "username", connConfig.Username)
		return nil, err
	}

	conn := new(Client)
	conn.client = c

	return conn, err
}

func (conn *Client) Disconnect() error {

	if conn.client == nil {
		// no connection
		return nil
	}
	return conn.client.Logout()
}
