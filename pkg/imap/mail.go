package imap

import (
	"bytes"
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	imapUtil "github.com/emersion/go-imap"
	imapMoveUtil "github.com/emersion/go-imap-move"
	imapClientUtil "github.com/emersion/go-imap/client"
	mailUtil "github.com/emersion/go-message/mail"
	"os"
	"strings"

	"time"
)

// System message flags, defined in RFC 3501 section 2.3.2.
const (
	SeenFlag     = "\\Seen"
	AnsweredFlag = "\\Answered"
	FlaggedFlag  = "\\Flagged"
	DeletedFlag  = "\\Deleted"
	DraftFlag    = "\\Draft"
	RecentFlag   = "\\Recent"
)

func UploadMails(c *imapClientUtil.Client, file string, mailbox string, flags []string) error {
	data, err := os.Open(file)
	defer data.Close()

	if err != nil {
		return err
	}

	msg := bytes.NewBuffer(nil)

	if _, err = msg.ReadFrom(data); err != nil {
		return err
	}

	// Select mailbox
	if _, err = Select(c, mailbox, false, true); err != nil {
		return err
	}

	return c.Append(mailbox, flags, time.Now(), msg)
}

func SearchMails(c *imapClientUtil.Client, mailbox string, withFlags []string, withoutFlags []string) ([]uint32, error) {

	// Select mailbox
	if _, err := Select(c, mailbox, true, false); err != nil {
		return nil, err
	}

	// Define search criteria
	criteria := imapUtil.NewSearchCriteria()
	if len(withFlags) > 0 {
		criteria.WithFlags = withFlags
	}
	if len(withoutFlags) > 0 {
		criteria.WithoutFlags = withoutFlags
	}

	// Actually search
	return c.UidSearch(criteria)
}

func FetchMails(c *imapClientUtil.Client, mailbox string, uids []uint32) ([]config.Mail, error) {

	// Select mailbox
	if _, err := Select(c, mailbox, true, false); err != nil {
		return nil, err
	}

	var fetchedMails []config.Mail

	seqset := imapUtil.SeqSet{}
	for _, uid := range uids {
		seqset.AddNum(uid)
	}

	var section imapUtil.BodySectionName
	section.Specifier = imapUtil.HeaderSpecifier // Loads all headers only (no body)
	items := []imapUtil.FetchItem{section.FetchItem(), imapUtil.FetchUid, imapUtil.FetchEnvelope}

	imapMessages := make(chan *imapUtil.Message, len(uids))
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(&seqset, items, imapMessages)
	}()

	var err error
	if err = <-done; err != nil {
		return nil, err
	}

	for imapMessage := range imapMessages {
		parsedHeaders, err := parseMailHeaders(imapMessage)
		if err != nil {
			return nil, err
		}
		fetchedMails = append(fetchedMails, config.NewMail(imapMessage, parsedHeaders))
	}

	return fetchedMails, err
}

func SearchAndFetchMails(c *imapClientUtil.Client, mailbox string, withFlags []string, withoutFlags []string) ([]config.Mail, error) {
	uids, err := SearchMails(c, mailbox, withFlags, withoutFlags)

	if err != nil || len(uids) == 0 {
		return nil, err
	}

	return FetchMails(c, mailbox, uids)
}

func DeleteMails(c *imapClientUtil.Client, mailbox string, uids []uint32, expunge bool) error {
	return SetMailFlags(c, mailbox, uids, "+FLAGS", []interface{}{imapUtil.DeletedFlag}, expunge)
}

func SetMailFlags(c *imapClientUtil.Client, mailbox string, uids []uint32, flagOp string, flags []interface{}, expunge bool) error {

	// Select mailbox
	if _, err := Select(c, mailbox, false, false); err != nil {
		return err
	}

	seqset := imapUtil.SeqSet{}
	for _, uid := range uids {
		seqset.AddNum(uid)
	}

	item := imapUtil.FormatFlagsOp(imapUtil.FlagsOp(flagOp), true)

	if err := c.UidStore(&seqset, item, flags, nil); err != nil {
		return err
	}

	if expunge {
		return c.Expunge(nil) //TODO set & verify ch to check list of expunged mails
	}

	return nil
}

func GetMailFlags(c *imapClientUtil.Client, mailbox string, uid uint32) ([]string, error) {
	var flags []string
	var err error

	// Select mailbox
	if _, err := Select(c, mailbox, true, false); err != nil {
		return nil, err
	}

	seqset := imapUtil.SeqSet{}
	seqset.AddNum(uid)

	items := []imapUtil.FetchItem{imapUtil.FetchFlags}

	imapMessages := make(chan *imapUtil.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(&seqset, items, imapMessages)
	}()

	if err = <-done; err != nil {
		return nil, err
	}

	for msg := range imapMessages {
		flags = msg.Flags
	}

	return flags, err
}

// List mailboxes
//mailboxes := make(chan *imapUtil.MailboxInfo, 11)
//done := make(chan error, 1)
//go func() {
//	done <- c.List("", "*", mailboxes)
//}()

//log.Println("Mailboxes:")
//for m := range mailboxes {
//	log.Println("* " + m.Name)
//}

func CreateMailbox(c *imapClientUtil.Client, name string) error {
	return c.Create(name)
}

//func MoveMail(acc *config.Account, mailbox string, uid uint32) error {
//	// Move BY COPYing and Deleting it
//	var err error
//	seqset := imapUtil.SeqSet{}
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

func MoveMails(c *imapClientUtil.Client, uids []uint32, from string, to string) error {
	var err error

	seqset := imapUtil.SeqSet{}
	for _, uid := range uids {
		seqset.AddNum(uid)
	}

	// Select mailbox
	if _, err := Select(c, from, false, false); err != nil {
		return err
	}

	moveClient := imapMoveUtil.NewClient(c)

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

func parseMailHeaders(rawMessage *imapUtil.Message) (config.MailHeaders, error) { //make private?
	headers := config.MailHeaders{}
	var err error

	// Create for mail parsing
	var section imapUtil.BodySectionName
	section.Specifier = imapUtil.HeaderSpecifier // Loads all headers only (no body)

	msgBody := rawMessage.GetBody(&section)
	if msgBody == nil {
		return headers, fmt.Errorf("server didn't returned message body for mail")
	}

	mr, err := mailUtil.CreateReader(msgBody)
	if err != nil {
		return headers, err
	}

	// Address Lists in headers
	addrFields := []string{"from", "to", "cc", "reply-to"}
	for _, fieldName := range addrFields {
		parsedList, err := parseAddrList(mr, fieldName, mr.Header.Get(fieldName))

		if err != nil {
			return nil, err
		} else {
			headers[fieldName] = parsedList
		}
	}

	// Some other standard envelope headers
	headers["subject"] = strings.ToLower(fmt.Sprintf("%v", rawMessage.Envelope.Subject))
	headers["date"] = strings.ToLower(fmt.Sprintf("%v", rawMessage.Envelope.Date))
	headers["message-id"] = strings.ToLower(fmt.Sprintf("%v", rawMessage.Envelope.MessageId))

	// All the other headers
	alreadyHandled := []string{"subject", "date", "message-id"}
	alreadyHandled = append(alreadyHandled, addrFields...)
	fields := mr.Header.Fields()
	for {
		next := fields.Next()
		if !next {
			break
		}

		fieldName := strings.ToLower(fields.Key())
		fieldValue := strings.ToLower(fields.Value())

		if contains(alreadyHandled, fieldName) {
			// we maintain these headers elsewhere
			continue
		}

		switch val := headers[fieldName].(type) {
		case nil:
			// detected new header
			headers[fieldName] = fieldValue
		case string:
			headerList := []string{val, fieldValue}
			headers[fieldName] = headerList
		case []string:
			headers[fieldName] = append(val, fieldValue)
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

	return headers, err
}

func parseAddrList(mr *mailUtil.Reader, fieldName string, fallback string) (string, error) {
	var fieldValue string
	addrs, err := mr.Header.AddressList(fieldName)

	if addrs == nil {
		//fmt.Println("yo", fieldName)
		// parsing failed, so return own or externally set fallback
		f := mr.Header.FieldsByKey(fieldName)
		if !f.Next() {
			return "", err
		} else {
			return fallback, nil
		}
	}

	if err != nil && err.Error() != "mail: missing '@' or angle-addr" { //ignore bad formated addrs
		// oh, real error
		return "", err
	}

	for _, addr := range addrs {
		formattedAddr := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v <%v>", addr.Name, addr.Address)))
		if fieldValue != "" {
			fieldValue += ", "
		}
		fieldValue += formattedAddr
	}

	return fieldValue, err
}

func contains(s []string, e string) bool { //TODO
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func Select(c *imapClientUtil.Client, mailbox string, readOnly bool, autoCreate bool) (*imapUtil.MailboxStatus, error) {
	status, err := c.Select(mailbox, readOnly)

	if err == nil {
		return status, err
	}

	// Select Failed, autocreate?
	if !autoCreate {
		return status, err
	}

	// Yes create and try SELECT again!
	if strings.HasPrefix(err.Error(), fmt.Sprintf("Mailbox doesn't exist: %v", mailbox)) {
		// SELECT failed because the target to did not exist. Create it and try again.
		if err = CreateMailbox(c, mailbox); err != nil {
			return nil, err
		}

		return Select(c, mailbox, readOnly, false)
	}

	return status, err
}
