package mail

import (
	"bytes"
	"fmt"
	"github.com/emersion/go-imap"
	imapMove "github.com/emersion/go-imap-move"
	imapClient "github.com/emersion/go-imap/client"
	"os"
	"strings"

	"time"
)

func UploadMails(c *imapClient.Client, file string, mailbox string, flags []string) error {
	data, err := os.Open(file)
	defer data.Close()

	if err != nil {
		return err
	}

	msg := bytes.NewBuffer(nil)

	if _, err = msg.ReadFrom(data); err != nil {
		return err
	}

	return c.Append(mailbox, flags, time.Now(), msg)
}

func SearchMails(c *imapClient.Client, mailbox string, withFlags []string, withoutFlags []string) ([]uint32, error) {

	// Select mailbox
	_, err := c.Select(mailbox, true)
	if err != nil {
		return []uint32{}, err
	}

	// Define search criteria
	criteria := imap.NewSearchCriteria()
	if len(withFlags) > 0 {
		criteria.WithFlags = withFlags
	}
	if len(withoutFlags) > 0 {
		criteria.WithoutFlags = withoutFlags
	}

	// Actually search
	return c.UidSearch(criteria)
}

func FetchMails(c *imapClient.Client, mailbox string, uids []uint32) ([]*imap.Message, error) {

	// Select mailbox
	_, err := c.Select(mailbox, true)
	if err != nil {
		return nil, err
	}

	var fetchedMails []*imap.Message

	seqset := imap.SeqSet{}
	for _, uid := range uids {
		seqset.AddNum(uid)
	}

	var section imap.BodySectionName
	section.Specifier = imap.HeaderSpecifier // Loads all headers only (no body)
	items := []imap.FetchItem{section.FetchItem(), imap.FetchUid}

	imapMessages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(&seqset, items, imapMessages)
	}()

	if err = <-done; err != nil {
		return fetchedMails, err
	}

	for imapMessage := range imapMessages {
		fetchedMails = append(fetchedMails, imapMessage)
	}

	return fetchedMails, err
}

func SearchAndFetchMails(c *imapClient.Client, mailbox string, withFlags []string, withoutFlags []string) ([]*imap.Message, error) {
	uids, err := SearchMails(c, mailbox, withFlags, withoutFlags)

	if err != nil || len(uids) == 0 {
		return nil, err
	}

	return FetchMails(c, mailbox, uids)

	/*
		msgBody := msg.GetBody(&section)

		if msgBody == nil {
			log.Fatal("Server didn't returned message body")
			return fetchedMails, nil
		}

		m, err := mail.CreateReader(msgBody)
		if err != nil {
			log.Fatal(err)
		}
	*/

	//fields := m.Header.FieldsByKey("received")
	//for {
	//	next := fields.Next()
	//	if !next { break }
	//	log.Println(fields.Key(), " => ", fields.Value())
	//
	//}

	/*
		date, _ := m.Header.Date()
		sub, _ := m.Header.Subject()
		from, _ := m.Header.AddressList("from")


		log.Println(date.Local(), sub, from[0].Name, )
	*/

	/*
		// Process each message's part
		for {
			p, err := m.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			switch h := p.Header.(type) {
			case *mail.InlineHeader:
				// This is the message's text (can be plain-text or HTML)
				b, _ := ioutil.ReadAll(p.Body)
				log.Printf("Got text: %v", string(b))
			case *mail.AttachmentHeader:
				// This is an attachment
				filename, _ := h.Filename()
				log.Printf("Got attachment: %v", filename)
			}

		}
	*/

	//log.Println("* " + msg.Envelope.MessageId + " / " + msg.Envelope.From[0].MailboxName, msg)
	//raw := msg.GetBody(section)
	//m, _ := mail.ReadMessage(raw)
	//log.Println(m.Header.Get("Received"))
	//	}

}

func DeleteMails(c *imapClient.Client, mailbox string, uids []uint32, expunge bool) error {
	return SetMailFlags(c, mailbox, uids, "+FLAGS", []interface{}{imap.DeletedFlag}, expunge)
}

func SetMailFlags(c *imapClient.Client, mailbox string, uids []uint32, flagOp string, flags []interface{}, expunge bool) error {

	if _, err := c.Select(mailbox, false); err != nil {
		return err
	}

	seqset := imap.SeqSet{}
	for _, uid := range uids {
		seqset.AddNum(uid)
	}

	item := imap.FormatFlagsOp(imap.FlagsOp(flagOp), true)

	if err := c.UidStore(&seqset, item, flags, nil); err != nil {
		return err
	}

	if expunge {
		return c.Expunge(nil) //TODO set & verify ch to check list of expunged mails
	}

	return nil
}

func GetMailFlags(c *imapClient.Client, mailbox string, uid uint32) ([]string, error) {
	var flags []string
	var err error

	if _, err := c.Select(mailbox, false); err != nil {
		return flags, err
	}

	seqset := imap.SeqSet{}
	seqset.AddNum(uid)

	items := []imap.FetchItem{imap.FetchFlags}

	imapMessages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(&seqset, items, imapMessages)
	}()

	if err = <-done; err != nil {
		return flags, err
	}

	for msg := range imapMessages {
		flags = msg.Flags
	}

	return flags, err
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

func CreateMailbox(c *imapClient.Client, name string) error {
	return c.Create(name)
}

//func MoveMail(acc *config.Account, mailbox string, uid uint32) error {
//	// Move BY COPYing and Deleting it
//	var err error
//	seqset := imap.SeqSet{}
//	seqset.AddNum(uid)
//
//	if err := acc.Connection.Client.Copy(&seqset, mailbox); err != nil {
//		if strings.HasPrefix(err.Error(), fmt.Sprintf("Mailbox doesn't exist: %v", mailbox)) {
//			// COPY failed becuase the target mailbox doesn't exist. Create it.
//			if err := CreateMailbox(acc, mailbox); err != nil {
//				return err
//			}
//
//			// Now retry COPY
//			if err := acc.Connection.Client.Copy(&seqset, mailbox); err != nil {
//				return err
//			}
//		}
//	}
//
//	// COPY to the new target mailbox seems to be successful. We can delete the mail from the old mailbox.
//	if err := DeleteMail(acc, mailbox, uid); err != nil {
//		return err
//	}
//
//	return err
//}

func MoveMails(c *imapClient.Client, uids []uint32, from string, to string) error {
	var err error

	seqset := imap.SeqSet{}
	for _, uid := range uids {
		seqset.AddNum(uid)
	}

	if _, err := c.Select(from, false); err != nil {
		return err
	}

	moveClient := imapMove.NewClient(c)

	err = moveClient.UidMove(&seqset, to)

	if err != nil && strings.HasPrefix(err.Error(), fmt.Sprintf("Mailbox doesn't exist: %v", to)) {
		// MOVE failed because the target to did not exist. Create it and try again.
		if err = CreateMailbox(c, to); err != nil {
			return err
		}

		if err = moveClient.UidMove(&seqset, to); err != nil {
			return err
		}
	}

	return err
}
