package filter

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/imap"
	imapClient "github.com/emersion/go-imap/client"
)

func GetUnsortedMails(c *imapClient.Client, inputMailbox config.InputMailbox) ([]config.Mail, error) {
	return imap.SearchAndFetchMails(c, inputMailbox.Mailbox, nil, inputMailbox.WithoutFlags)
}

func EvaluateFilterSetsOnMails(acc config.Account) ([]config.Mail, error) {

	var remainingMails []config.Mail
	msgs, err := GetUnsortedMails(acc.Connection.Client, *acc.InputMailbox)

	for _, msg := range msgs {
		var matched bool
		for _, filterSet := range acc.FilterSet {
			matched, err = ParseRuleSet(filterSet.RuleSet, msg.Headers)

			if err != nil {
				return nil, err
			}

			if matched {
				err = RunCommands(acc.Connection.Client, acc.InputMailbox.Mailbox, msg.RawMail.Uid, filterSet.Commands)
				if err != nil {
					return nil, err //TODO
				}
			}
		}

		if !matched {
			remainingMails = append(remainingMails, msg)
		}
	}

	for _, msg := range remainingMails {
		if *acc.FallbackMailbox == acc.InputMailbox.Mailbox || *acc.FallbackMailbox == "" {
			err = imap.SetMailFlags(acc.Connection.Client, acc.InputMailbox.Mailbox, []uint32{msg.RawMail.Uid}, "+FLAGS", []interface{}{imap.FlaggedFlag}, false)
			if err != nil {
				return nil, err //TODO
			}
		} else {
			err = imap.MoveMails(acc.Connection.Client, []uint32{msg.RawMail.Uid}, acc.InputMailbox.Mailbox, *acc.FallbackMailbox)
			if err != nil {
				return nil, err //TODO
			}
		}
	}

	return remainingMails, err
}
