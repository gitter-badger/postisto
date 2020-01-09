package filter

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/imap"
	"github.com/arnisoph/postisto/pkg/log"
)

func GetUnsortedMsgs(imapClient *imap.Client, inputMailbox config.InputMailbox) ([]*imap.Message, error) {
	return imapClient.SearchAndFetch(inputMailbox.Mailbox, nil, inputMailbox.WithoutFlags)
}

func EvaluateFilterSetsOnMsgs(imapClient *imap.Client, acc config.Account) error {

	var remainingMsgs []*imap.Message
	msgs, err := GetUnsortedMsgs(imapClient, *acc.InputMailbox)

	for _, msg := range msgs {
		var matched bool

		log.Debugw("Starting to filter message", "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId, "headers", msg.Headers)
		for _, filterSet := range acc.FilterSet {
			log.Debugw("Evaluate rule set against message headers", "uid", msg.RawMessage.Uid, "ruleSet", filterSet.RuleSet)
			matched, err = ParseRuleSet(filterSet.RuleSet, msg.Headers)

			if err != nil {
				return err
			}

			if !matched {
				continue
			}

			log.Infow("IT'S A MATCH! Apply commands to message via IMAP..", "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId, "cmd", filterSet.Commands)
			err = RunCommands(imapClient, acc.InputMailbox.Mailbox, msg.RawMessage.Uid, filterSet.Commands)
			if err != nil {
				log.Errorw("Failed to run command on matched message", err, "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId, "cmd", filterSet.Commands)
				return err
			}

			break
		}

		if !matched {
			log.Debugw("No filter matched to this message, scheduling fallback action (flag/move)", "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId, "headers", msg.Headers)
			remainingMsgs = append(remainingMsgs, msg)
		}
	}

	for _, msg := range remainingMsgs {
		if *acc.FallbackMailbox == acc.InputMailbox.Mailbox || *acc.FallbackMailbox == "" {
			log.Infow("No filter matched to this message. Flagging the message now.", "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId)
			if err := imapClient.SetFlags(acc.InputMailbox.Mailbox, []uint32{msg.RawMessage.Uid}, "+FLAGS", []interface{}{imap.FlaggedFlag}, false); err != nil {
				return err
			}
		} else {
			log.Infow("No filter matched to this message. Moving it to the fallback mailbox now.", "uid", msg.RawMessage.Uid, "message_id", msg.RawMessage.Envelope.MessageId, "mailbox", acc.FallbackMailbox)
			if err := imapClient.Move([]uint32{msg.RawMessage.Uid}, acc.InputMailbox.Mailbox, *acc.FallbackMailbox); err != nil {
				return err
			}
		}
	}

	return nil
}
