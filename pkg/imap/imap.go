package imap

import (
	"bytes"
	"fmt"
	imapUtil "github.com/emersion/go-imap"
	imapMoveUtil "github.com/emersion/go-imap-move"
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

func (conn *Client) Upload(file string, mailbox string, flags []string) error {
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
	if _, err = conn.Select(mailbox, false, true); err != nil {
		return err
	}

	return conn.client.Append(mailbox, flags, time.Now(), msg)
}

func (conn *Client) Search(mailbox string, withFlags []string, withoutFlags []string) ([]uint32, error) {

	// Select mailbox
	if _, err := conn.Select(mailbox, true, false); err != nil {
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
	return conn.client.UidSearch(criteria)
}

func (conn *Client) Fetch(mailbox string, uids []uint32) ([]*Message, error) {

	// Select mailbox
	if _, err := conn.Select(mailbox, true, false); err != nil {
		return nil, err
	}

	var fetchedMails []*Message

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
		done <- conn.client.UidFetch(&seqset, items, imapMessages)
	}()

	var err error
	if err = <-done; err != nil {
		return nil, err
	}

	for imapMessage := range imapMessages {
		parsedHeaders, err := parseMessageHeaders(imapMessage)
		if err != nil {
			return nil, err
		}
		fetchedMails = append(fetchedMails, NewMessage(imapMessage, parsedHeaders))
	}

	return fetchedMails, err
}

func (conn *Client) SearchAndFetch(mailbox string, withFlags []string, withoutFlags []string) ([]*Message, error) {
	uids, err := conn.Search(mailbox, withFlags, withoutFlags)

	if err != nil || len(uids) == 0 {
		return nil, err
	}

	return conn.Fetch(mailbox, uids)
}

func (conn *Client) Delete(mailbox string, uids []uint32, expunge bool) error {
	return conn.SetFlags(mailbox, uids, "+FLAGS", []interface{}{imapUtil.DeletedFlag}, expunge)
}

func (conn *Client) SetFlags(mailbox string, uids []uint32, flagOp string, flags []interface{}, expunge bool) error {

	// Select mailbox
	if _, err := conn.Select(mailbox, false, false); err != nil {
		return err
	}

	seqset := imapUtil.SeqSet{}
	for _, uid := range uids {
		seqset.AddNum(uid)
	}

	item := imapUtil.FormatFlagsOp(imapUtil.FlagsOp(flagOp), true)

	if err := conn.client.UidStore(&seqset, item, flags, nil); err != nil {
		return err
	}

	if expunge {
		return conn.client.Expunge(nil) //TODO set & verify ch to check list of expunged mails
	}

	return nil
}

func (conn *Client) GetFlags(mailbox string, uid uint32) ([]string, error) {
	var flags []string
	var err error

	// Select mailbox
	if _, err := conn.Select(mailbox, true, false); err != nil {
		return nil, err
	}

	seqset := imapUtil.SeqSet{}
	seqset.AddNum(uid)

	items := []imapUtil.FetchItem{imapUtil.FetchFlags}

	imapMessages := make(chan *imapUtil.Message, 1)
	done := make(chan error, 1)
	go func() {
		done <- conn.client.UidFetch(&seqset, items, imapMessages)
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

func (conn *Client) CreateMailbox(name string) error {
	return conn.client.Create(name)
}

//func MoveMail(acc *config.Account, mailbox string, uid uint32) error {
//	// Move BY COPYing and Deleting it
//	var err error
//	seqset := imapUtil.SeqSet{}
//	seqset.AddNum(uid)
//
//	if err := acc.Client.Client.Copy(&seqset, mailbox); err != nil {
//		if strings.HasPrefix(err.Error(), fmt.Sprintf("Mailbox doesn't exist: %v", mailbox)) {
//			// COPY failed becuase the target mailbox doesn't exist. Create it.
//			if err := CreateMailbox(acc, mailbox); err != nil {
//				return err
//			}
//
//			// Now retry COPY
//			if err := acc.Client.Client.Copy(&seqset, mailbox); err != nil {
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

func (conn *Client) Move(uids []uint32, from string, to string) error {
	var err error

	seqset := imapUtil.SeqSet{}
	for _, uid := range uids {
		seqset.AddNum(uid)
	}

	// Select mailbox
	if _, err := conn.Select(from, false, false); err != nil {
		return err
	}

	moveClient := imapMoveUtil.NewClient(conn.client)

	err = moveClient.UidMove(&seqset, to)

	if err != nil && strings.HasPrefix(err.Error(), fmt.Sprintf("Mailbox doesn't exist: %v", to)) {
		// MOVE failed because the target to did not exist. Create it and try again.
		if err = conn.CreateMailbox(to); err != nil {
			return err
		}

		if err = moveClient.UidMove(&seqset, to); err != nil {
			return err
		}
	}

	return err
}

func (conn *Client) Select(mailbox string, readOnly bool, autoCreate bool) (*imapUtil.MailboxStatus, error) {
	status, err := conn.client.Select(mailbox, readOnly)

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
		if err = conn.CreateMailbox(mailbox); err != nil {
			return nil, err
		}

		return conn.Select(mailbox, readOnly, false)
	}

	return status, err
}
