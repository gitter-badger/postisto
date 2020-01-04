package mail

import (
	"bytes"
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap-move"
	"os"
	"strings"

	"time"
)

func UploadMails(acc *config.Account, file string, mailbox string, flags []string) error {
	data, err := os.Open(file)
	defer data.Close()

	if err != nil {
		return err
	}

	msg := bytes.NewBuffer(nil)

	if _, err = msg.ReadFrom(data); err != nil {
		return err
	}

	return acc.Connection.Client.Append(mailbox, flags, time.Now(), msg)
}

func searchMails(acc *config.Account) ([]uint32, error) {

	// Select mailbox
	_, err := acc.Connection.Client.Select(acc.Connection.InputMailbox.Mailbox, true)
	if err != nil {
		return []uint32{}, err
	}

	// Define search criteria
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = acc.Connection.InputMailbox.WithoutFlags

	// Actually search
	return acc.Connection.Client.UidSearch(criteria)
}

func SearchAndFetchMails(acc *config.Account) ([]*imap.Message, error) {
	var fetchedMails []*imap.Message
	uids, err := searchMails(acc)

	if err != nil || len(uids) == 0 {
		return fetchedMails, nil
	}

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
		err := acc.Connection.Client.Fetch(&seqset, items, imapMessages)
		done <- err
	}()

	for msg := range imapMessages {
		fetchedMails = append(fetchedMails, msg)

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
	}

	return fetchedMails, err
}

func DeleteMail(acc *config.Account, mailbox string, uid uint32) error {
	return SetMailFlags(acc, mailbox, uid, "+FLAGS", []interface{}{imap.DeletedFlag})
}

func SetMailFlags(acc *config.Account, mailbox string, uid uint32, flagOp string, flags []interface{}) error {

	if _, err := acc.Connection.Client.Select(mailbox, false); err != nil {
		return err
	}

	seqset := imap.SeqSet{}
	seqset.AddNum(uid)

	item := imap.FormatFlagsOp(imap.FlagsOp(flagOp), true)

	return acc.Connection.Client.UidStore(&seqset, item, flags, nil)
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

func CreateMailbox(acc *config.Account, name string) error {
	return acc.Connection.Client.Create(name)
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

func MoveMail(acc *config.Account, uid uint32, mailbox string) error {
	var err error
	seqset := imap.SeqSet{}
	seqset.AddNum(uid)

	moveClient := move.NewClient(acc.Connection.Client)

	err = moveClient.UidMove(&seqset, mailbox)

	if err != nil && strings.HasPrefix(err.Error(), fmt.Sprintf("Mailbox doesn't exist: %v", mailbox)) {
		// MOVE failed because the target mailbox did not exist. Create it and try again.
		if err = CreateMailbox(acc, mailbox); err != nil {
			return err
		}

		if err = moveClient.UidMove(&seqset, mailbox); err != nil {
			return err
		}
	}

	return err
}
