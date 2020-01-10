package main

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/filter"
	"github.com/arnisoph/postisto/pkg/log"
	"time"
)

func main() {

	configPath := "/Users/ab/Documents/dev/GOPATH/src/github.com/arnisoph/postisto/test/data/configs/examples/basic/"
	cfg, err := config.NewConfigFromFile(configPath)
	if err != nil {
		log.Fatalw("Failed to load configuration", err, "path", configPath)
	}

	if err := log.InitWithConfig(cfg.Settings.LogConfig); err != nil {
		log.Fatalw("Failed to set up logging", err)
	}

	acc := cfg.Accounts["gmail"]
	filters := cfg.Filters["gmail"]

	if err := acc.Connection.Connect(); err != nil {
		log.Fatalw("Failed to login to IMAP server", err)
	}

	log.Infow("=>", "inbox", acc.InputMailbox, "fallback", acc.FallbackMailbox)

	for {

		if err := filter.EvaluateFilterSetsOnMsgs(&acc.Connection, acc.InputMailbox.Mailbox, nil, *acc.FallbackMailbox, filters); err != nil {
			log.Fatal("Failed to run filter engine", err)
		}

		time.Sleep(time.Second * 10)
	}

	//// NewConfigFromFile user config
	//var err error
	//cfg := config.New()
	//if cfg, err = cfg.NewConfigFromFile("/Users/ab/Documents/dev/GOPATH/src/github.com/arnisoph/postisto/test/data/configs/valid/"); err != nil {
	//	log.Panicf("failed to load config: %v", err)
	//}
	//
	//// Connect to IMAP servers
	//accs := map[string]*config.Account{}
	//for accName, acc := range cfg.Accounts {
	//	if !acc.Connection.Enabled {
	//		continue
	//	}
	//	if err := conn.Connect(acc); err != nil {
	//		log.Fatalf("failed to connect (%v): %v", accName, err)
	//
	//	} else {
	//		accs[accName] = acc
	//	}
	//}
	//
	//defer func() {
	//	for _, acc := range accs {
	//		if err := conn.Disconnect(acc); err != nil {
	//			log.Fatalf("failed to discoonect account %v", err)
	//		}
	//	}
	//}()
	//
	//for _, acc := range accs {
	//	log.Println(acc.Connection.Connection.State(), imap.AuthenticatedState)
	//}
	//
	//log.Println("Done!")
}
