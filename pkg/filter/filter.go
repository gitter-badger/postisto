package filter

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/mail"
	imapClient "github.com/emersion/go-imap/client"
)

func getUnsortedMails(c *imapClient.Client, inputMailbox config.InputMailbox) ([]config.Mail, error) {
	return mail.SearchAndFetchMails(c, inputMailbox.Mailbox, nil, inputMailbox.WithoutFlags)
}

func EvaluateFilterSetOnMails(acc config.Account) (bool, error) {

	mails, err := getUnsortedMails(acc.Connection.Client, *acc.Connection.InputMailbox)

	for _, mail := range mails {
		var matched bool
		for filterName, filterSet := range acc.FilterSet {
			matched, err = ParseRuleSet(filterSet.RuleSet, config.MailHeaders{"from": "foo@example.com", "to": "me@EXAMPLE.com", "subject": "With Löve", "empty-header": ""})
			if err != nil {
				return false, err
			}

			if matched {
				return true, err
			}
			fmt.Println(filterName, matched, mail)
		}
	}

	return false, err
}
