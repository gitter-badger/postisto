package main

import (
	"bytes"
	"github.com/arnisoph/postisto/config"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"log"
	"os"
	"time"
)

func main() {
	log.Println("Connecting to server...")

	// Connect to server
	c, err := client.Dial("localhost:10143")

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")

	// Don't forget to logout
	defer c.Logout()

	// Login
	if err := c.Login("test-42@example.com", "test"); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 11)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		log.Println("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	//if err := <-done; err != nil {
	//	log.Fatal(err)
	//}

	//log.Println("Creating new mailboxes..")

	//if err := c.Create("test123"); err != nil {
	//	log.Fatal(err)
	//}

	data, err := os.Open("/Users/ab/Documents/dev/github/tabellarius/tests/mails/log1.txt")
	if err != nil {
		log.Fatalf("-> %v", err)
	}
	defer data.Close()

	b := bytes.NewBuffer(nil)
	b.ReadFrom(data)

	if err := c.Append("test123", []string{}, time.Now(), b); err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	mbox, err := c.Select("test123", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Flags for test123:", mbox.Flags)

	// Get the last 4 messages
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > 3 {
		// We're using unsigned integers here, only substract if the result is > 0
		from = mbox.Messages - 3
	}
	seqset := new(imap.SeqSet)
	seqset.AddRange(from, to)

	messages := make(chan *imap.Message, 10)
	done = make(chan error, 1)
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope}, messages)
	}()

	log.Println("Last 4 messages:")
	for msg := range messages {
		log.Println("* " + msg.Envelope.MessageId + " / " + msg.Envelope.From[0].MailboxName)
	}

	log.Printf("%v - %v", mbox.Items, mbox.Messages)

	log.Println("Done!")

	cfg, err := config.GetConfig("/Users/ab/Documents/dev/GOPATH/src/github.com/arnisoph/tabellarius2/tests/configs/valid")
	log.Println(err, cfg.Filters["test"]["test_unicode_from"].Rules)
}
