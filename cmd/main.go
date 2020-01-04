package main

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/conn"
	"github.com/emersion/go-imap"
	"log"
)

func main() {
	// Load user config
	var err error
	cfg := config.New()
	if cfg, err = cfg.Load("/Users/ab/Documents/dev/GOPATH/src/github.com/arnisoph/postisto/test/data/configs/valid/"); err != nil {
		log.Panicf("failed to load config: %v", err)
	}

	// Connect to IMAP servers
	accs := map[string]*config.Account{}
	for accName, acc := range cfg.Accounts {
		if !acc.Connection.Enabled {
			continue
		}
		if err := conn.Connect(acc); err != nil {
			log.Fatalf("failed to connect (%v): %v", accName, err)

		} else {
			accs[accName] = acc
		}
	}

	defer func() {
		for _, acc := range accs {
			if err := conn.Disconnect(acc); err != nil {
				log.Fatalf("failed to discoonect account %v", err)
			}
		}
	}()

	for _, acc := range accs {
		log.Println(acc.Connection.Client.State(), imap.AuthenticatedState)
	}

	log.Println("Done!")
}
