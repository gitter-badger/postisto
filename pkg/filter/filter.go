package filter

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/imap"
	"github.com/arnisoph/postisto/pkg/log"
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
			log.Debugw("Starting to evaluate rule set against message headers", "ruleSet", filterSet.RuleSet, "headers", msg.Headers)
			matched, err = ParseRuleSet(filterSet.RuleSet, msg.Headers)

			if err != nil {
				return nil, err
			}

			if matched {
				log.Infow("IT'S A MATCH! Apply commands to message via IMAP..", "message_id", msg.RawMessage.Envelope.MessageId, "cmd", filterSet.Commands)
				err = RunCommands(imapClient, acc.InputMailbox.Mailbox, msg.RawMessage.Uid, filterSet.Commands)
				if err != nil {
					return nil, err
				}
			}
		}

		if !matched {
			remainingMsgs = append(remainingMsgs, msg)
		}
	}

	for _, msg := range remainingMsgs {
		if *acc.FallbackMailbox == acc.InputMailbox.Mailbox || *acc.FallbackMailbox == "" {
			log.Infow("No filter matched to this message. Moving it to the fallbock mailbox.", "mailbox", acc.FallbackMailbox)
			if err = imapClient.SetFlags(acc.InputMailbox.Mailbox, []uint32{msg.RawMessage.Uid}, "+FLAGS", []interface{}{imap.FlaggedFlag}, false); err != nil {
				return nil, err
			}
		} else {
			log.Infow("No filter matched to this message. Moving it to the fallbock mailbox.", "mailbox", acc.FallbackMailbox)
			if err = imapClient.Move([]uint32{msg.RawMessage.Uid}, acc.InputMailbox.Mailbox, *acc.FallbackMailbox); err != nil {
				return nil, err
			}
		}
	}

	return remainingMsgs, nil
}
