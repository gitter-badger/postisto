package filter

import (
	"github.com/emersion/go-imap"
	imapClient "github.com/emersion/go-imap/client"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/mail"
)

func GetUnsortedMails(c *imapClient.Client, inputMailbox config.InputMailbox) ([]*imap.Message, error) {
	return mail.SearchAndFetchMails(c, inputMailbox.Mailbox, nil, inputMailbox.WithoutFlags)
}
