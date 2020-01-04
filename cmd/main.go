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
	conns := map[string]*config.Account{}
	for accName, acc := range cfg.Accounts {
		if !acc.Connection.Enabled {
			continue
		}
		if err := conn.Connect(acc); err != nil {
			log.Fatalf("failed to connect (%v): %v", accName, err)

		} else {
			conns[accName] = acc
		}
	}

	defer func() {
		if err := conn.DisconnectAll(conns); err != nil {
			log.Panic("failed to discoonect account", err)
		}
	}()

	for _, acc := range conns {
		log.Println(acc.Connection.Client.State(), imap.AuthenticatedState)
	}

	// List mailboxes
	//mailboxes := make(chan *imap.MailboxInfo, 11)
	//done := make(chan error, 1)
	//go func() {
	//	done <- c.List("", "*", mailboxes)
	//}()

	//log.Println("Mailboxes:")
	//for m := range mailboxes {
	//	log.Println("* " + m.Name)
	//}

	//if err := <-done; err != nil {
	//	log.Fatal(err)
	//}

	//if err := <-done; err != nil {
	//	log.Fatal(err)
	//}

	//log.Println("Creating new mailboxes..")

	//if err := c.Create("test123"); err != nil {
	//	log.Fatal(err)
	//}

	//msg := <-messages
	//raw := msg.GetBody(section)
	//if raw == nil {
	//	log.Fatal("Server didn't returned message body")
	//}
	//
	//if err := <-done; err != nil {
	//	log.Fatal(err)
	//}
	//
	//m, err := mail.ReadMessage(raw)
	//if err != nil {
	//	log.Fatal(err)
	//}

	//for msg := range messages {
	//	r := msg.GetBody(&section)
	//
	//	if r == nil {
	//		log.Fatal("Server didn't returned message body")
	//	}
	//
	//	m, err := mail.CreateReader(r)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	fields := m.Header.FieldsByKey("received")
	//	for {
	//		next := fields.Next()
	//		if !next { break }
	//		log.Println(fields.Key(), " => ", fields.Value())
	//
	//	}
	//
	//	date, _ := m.Header.Date()
	//	sub, _ := m.Header.Subject()
	//	from, _ := m.Header.AddressList("from")
	//	log.Println(date.Local(), sub, from[0].Name, )
	//
	//	//log.Println("* " + msg.Envelope.MessageId + " / " + msg.Envelope.From[0].MailboxName, msg)
	//	//raw := msg.GetBody(section)
	//	//m, _ := mail.ReadMessage(raw)
	//	//log.Println(m.Header.Get("Received"))
	//	log.Println("============================================")
	//}

	//log.Printf("%v - %v", mbox.Items, mbox.Messages)

	log.Println("Done!")

}
