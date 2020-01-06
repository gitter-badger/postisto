package mail

import (
	"bytes"
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/emersion/go-imap"
	imapMove "github.com/emersion/go-imap-move"
	imapClient "github.com/emersion/go-imap/client"
	mailUtil "github.com/emersion/go-message/mail"
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

func FetchMails(c *imapClient.Client, mailbox string, uids []uint32) ([]config.Mail, error) {

	// Select mailbox
	_, err := c.Select(mailbox, true)
	if err != nil {
		return nil, err
	}

	var fetchedMails []config.Mail

	seqset := imap.SeqSet{}
	for _, uid := range uids {
		seqset.AddNum(uid)
	}

	var section imap.BodySectionName
	section.Specifier = imap.HeaderSpecifier // Loads all headers only (no body)
	items := []imap.FetchItem{section.FetchItem(), imap.FetchUid, imap.FetchEnvelope}

	imapMessages := make(chan *imap.Message, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(&seqset, items, imapMessages)
	}()

	if err = <-done; err != nil {
		return fetchedMails, err
	}

	for imapMessage := range imapMessages {
		fetchedMails = append(fetchedMails, config.NewMail(imapMessage))
	}

	return fetchedMails, err
}

func SearchAndFetchMails(c *imapClient.Client, mailbox string, withFlags []string, withoutFlags []string) ([]config.Mail, error) {
	uids, err := SearchMails(c, mailbox, withFlags, withoutFlags)

	if err != nil || len(uids) == 0 {
		return nil, err
	}

	return FetchMails(c, mailbox, uids)
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

func ParseMailHeaders(c *imapClient.Client, mails []config.Mail) ([]config.Mail, error) {
	var err error

	var section imap.BodySectionName
	section.Specifier = imap.HeaderSpecifier // Loads all headers only (no body)

	for _, msg := range mails {
		msgBody := msg.RawMail.GetBody(&section)

		if msgBody == nil {
			return mails, fmt.Errorf("server didn't returned message body for mail")
		}

		mr, err := mailUtil.CreateReader(msgBody)
		if err != nil {
			return mails, err
		}

		fields := mr.Header.Fields()
		//fields := m.Header.FieldsByKey("received")

		addrFields := []string{"from", "to", "cc"}
		for _, field := range addrFields {
			addrs, err := mr.Header.AddressList(field)
			if err != nil && err.Error() != "mail: missing '@' or angle-addr" { //ignore bad formated addrs
				return mails, err
			}

			for _, addr := range addrs {
				if msg.Headers[field] != "" {
					msg.Headers[field] += ", "
				}
				msg.Headers[field] += strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v <%v>", addr.Name, addr.Address)))
			}
		}

		msg.Headers["subject"] = strings.ToLower(fmt.Sprintf("%v", msg.RawMail.Envelope.Subject))
		msg.Headers["date"] = strings.ToLower(fmt.Sprintf("%v", msg.RawMail.Envelope.Date))
		msg.Headers["reply-to"] = strings.ToLower(fmt.Sprintf("%v", msg.RawMail.Envelope.ReplyTo))
		msg.Headers["message-id"] = strings.ToLower(fmt.Sprintf("%v", msg.RawMail.Envelope.MessageId))

		for {
			next := fields.Next()
			if !next {
				break
			}

			if msg.Headers[strings.ToLower(fields.Key())] == "" { //TODO support received?
				//fmt.Println("schreibne", fields.Key())
				msg.Headers[strings.ToLower(fields.Key())] = strings.ToLower(fields.Value())
			}
		}

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
	}

	return mails, err
}
