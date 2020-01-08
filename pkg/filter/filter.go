package filter

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/imap"
)

func GetUnsortedMsgs(imapClient *imap.Client, inputMailbox config.InputMailbox) ([]*imap.Message, error) {
	return imapClient.SearchAndFetch(inputMailbox.Mailbox, nil, inputMailbox.WithoutFlags)
}

func EvaluateFilterSetsOnMsgs(imapClient *imap.Client, acc config.Account) ([]*imap.Message, error) {

	var remainingMsgs []*imap.Message
	msgs, err := GetUnsortedMsgs(imapClient, *acc.InputMailbox)

	for _, msg := range msgs {
		var matched bool
		for _, filterSet := range acc.FilterSet {
			matched, err = ParseRuleSet(filterSet.RuleSet, msg.Headers)

			if err != nil {
				return nil, err
			}

			if matched {
				err = RunCommands(imapClient, acc.InputMailbox.Mailbox, msg.RawMessage.Uid, filterSet.Commands)
				if err != nil {
					return nil, err //TODO
				}
			}
		}

		if !matched {
			remainingMsgs = append(remainingMsgs, msg)
		}
	}

	for _, msg := range remainingMsgs {
		if *acc.FallbackMailbox == acc.InputMailbox.Mailbox || *acc.FallbackMailbox == "" {
			err = imapClient.SetFlags(acc.InputMailbox.Mailbox, []uint32{msg.RawMessage.Uid}, "+FLAGS", []interface{}{imap.FlaggedFlag}, false)
			if err != nil {
				return nil, err //TODO
			}
		} else {
			err = imapClient.Move([]uint32{msg.RawMessage.Uid}, acc.InputMailbox.Mailbox, *acc.FallbackMailbox)
			if err != nil {
				return nil, err //TODO
			}
		}
	}

	return remainingMsgs, err
}
