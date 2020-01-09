package imap

import (
	"bytes"
	"fmt"
	"github.com/arnisoph/postisto/pkg/log"
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
		log.Errorw("Failed to upload message to mailbox", err, "mailbox", mailbox)
		return err
	}

	msg := bytes.NewBuffer(nil)

	if _, err = msg.ReadFrom(data); err != nil {
		log.Errorw("Failed to upload message to mailbox", err, "mailbox", mailbox)
		return err
	}

	// Select mailbox
	if _, err = conn.Select(mailbox, false, true); err != nil {
		log.Errorw("Failed to upload message to mailbox", err, "mailbox", mailbox)
		return err
	}

	// Upload (APPEND)
	if err = conn.client.Append(mailbox, flags, time.Now(), msg); err != nil {
		log.Errorw("Failed to upload message to mailbox", err, "mailbox", mailbox)
		return err
	}

	return nil
}

func (conn *Client) Search(mailbox string, withFlags []string, withoutFlags []string) ([]uint32, error) {

	// Select mailbox
	if _, err := conn.Select(mailbox, true, false); err != nil {
		log.Errorw("Failed to open mailbox for searching", err, "mailbox", mailbox)
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
		log.Errorw("Failed to open mailbox to fetching messages", err, "mailbox", mailbox)
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
		log.Errorw("Failed to fetch message from mailbox", err, "mailbox", mailbox)
		return nil, err
	}

	for imapMessage := range imapMessages {
		parsedHeaders, err := parseMessageHeaders(imapMessage)
		if err != nil {
			log.Errorw("Failed to parse message headers", err, "mailbox", mailbox, "message_subject", imapMessage.Envelope.Subject, "message_id", imapMessage.Envelope.MessageId)
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

func (conn *Client) DeleteMsgs(mailbox string, uids []uint32, expunge bool) error {
	return conn.SetFlags(mailbox, uids, "+FLAGS", []interface{}{imapUtil.DeletedFlag}, expunge)
}

func (conn *Client) SetFlags(mailbox string, uids []uint32, flagOp string, flags []interface{}, expunge bool) error {

	// Select mailbox
	if _, err := conn.Select(mailbox, false, false); err != nil {
		log.Errorw("Failed to open mailbox to set message flags", err, "mailbox", mailbox)
		return err
	}

	seqset := imapUtil.SeqSet{}
	for _, uid := range uids {
		seqset.AddNum(uid)
	}

	item := imapUtil.FormatFlagsOp(imapUtil.FlagsOp(flagOp), true)

	if err := conn.client.UidStore(&seqset, item, flags, nil); err != nil {
		log.Errorw("Failed to set message flags", err, "mailbox", mailbox)
		return err
	}

	if expunge {
		if err := conn.client.Expunge(nil); err != nil {
			log.Errorw("Failed to expunge after setting message flags", err, "mailbox", mailbox)
			return err
		}
	}

	return nil
}

func (conn *Client) GetFlags(mailbox string, uid uint32) ([]string, error) {
	var flags []string
	var err error

	// Select mailbox
	if _, err := conn.Select(mailbox, true, false); err != nil {
		log.Errorw("Failed to open mailbox to get messages flags", err, "mailbox", mailbox)
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
		log.Errorw("Failed to fetch message from mailbox", err, "mailbox", mailbox)
		return nil, err
	}

	for msg := range imapMessages {
		flags = msg.Flags
	}

	return flags, nil
}

func (conn *Client) CreateMailbox(name string) error {
	log.Infow("Creating new mailbox", "mailbox", name)
	if err := conn.client.Create(name); err != nil {
		log.Errorw("Failed to create mailbox", err, "mailbox", name)
		return err
	}

	return nil
}

func (conn *Client) DeleteMailbox(name string) error {
	log.Infow("Deleting mailbox", "mailbox", name)
	if err := conn.client.Delete(name); err != nil {
		log.Errorw("Failed to delete mailbox ", err, "mailbox", name)
		return err
	}

	return nil
}

// List mailboxes
func (conn *Client) List() (map[string]imapUtil.MailboxInfo, error) {
	var err error

	mailboxesChan := make(chan *imapUtil.MailboxInfo, 100)
	done := make(chan error, 1)
	go func() {
		done <- conn.client.List("", "*", mailboxesChan)
	}()

	if err = <-done; err != nil {
		log.Error("Failed to list mailboxes", err)
		return nil, err
	}

	mailboxes := map[string]imapUtil.MailboxInfo{}
	for mailBox := range mailboxesChan {
		mailboxes[mailBox.Name] = *mailBox
	}

	return mailboxes, nil
}

type Queue struct {
	elements chan interface{}
}

func NewQueue(size int) *Queue {
	return &Queue{
		elements: make(chan interface{}, size),
	}
}

func (queue *Queue) Push(element interface{}) {
	select {
	case queue.elements <- element:
	default:
		panic("Queue full")
	}
}

func (queue *Queue) Pop() interface{} {
	select {
	case e := <-queue.elements:
		return e
	default:
		panic("Queue empty")
	}
	return nil
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
		log.Errorw("Failed to open mailbox to move messages", err, "source", from, "destination", to)
		return err
	}

	moveClient := imapMoveUtil.NewClient(conn.client)
	err = moveClient.UidMove(&seqset, to)

	if err == nil {
		return nil
	}

	// Move failed
	if strings.Contains(err.Error(), "Mailbox doesn't exist") ||
		strings.Contains(err.Error(), "No folder") {
		mailBoxes, err := conn.List()
		if err != nil {
			log.Errorw("Failed to move messages after trying to get list of mailboxes", err, "source", from, "destination", to)
			return err
		}

		if _, notFound := mailBoxes[to]; notFound == false {
			// MOVE failed because the target to did not exist. Create it and try again.
			if err := conn.CreateMailbox(to); err != nil {
				return err
			}

			return conn.Move(uids, from, to)
		}
	}

	log.Errorw("Failed to move messages for an unexpected reason", err, "source", from, "destination", to)
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
	if strings.HasPrefix(err.Error(), fmt.Sprintf("Mailbox doesn't exist: %v", mailbox)) ||
		strings.Contains(err.Error(), "Unknown Mailbox") {
		// SELECT failed because the target to did not exist. Create it and try again.
		if err = conn.CreateMailbox(mailbox); err != nil {
			return nil, err
		}

		return conn.Select(mailbox, readOnly, false)
	}

	return status, err
}
