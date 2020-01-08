package filter

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/imap"
	imapClient "github.com/emersion/go-imap/client"
)

//type UnknownCommandTypeError struct {
//	typeName string
//}
//
//func (err *UnknownCommandTypeError) Error() string {
//	return fmt.Sprintf("Command type %q unknown", err.typeName)
//}
//
//type BadCommandTargetError struct {
//	targetName string
//}
//
//func (err *BadCommandTargetError) Error() string {
//	return fmt.Sprintf("Bad command target %q", err.targetName)
//}

func RunCommands(c *imapClient.Client, from string, uid uint32, cmds config.FilterOps) error {
	var err error
	uids := []uint32{uid}

	if cmds["move"] != nil {
		if err := imap.MoveMails(c, uids, from, cmds["move"].(string)); err != nil {
			return err
		}
	}

	to := from
	if cmds["move"] != nil {
		to = cmds["move"].(string)
	}

	if cmds["add_flags"] != nil {
		if err := imap.SetMailFlags(c, to, uids, "+FLAGS", cmds["add_flags"].([]interface{}), false); err != nil {
			return err
		}
	}

	if cmds["remove_flags"] != nil {
		if err := imap.SetMailFlags(c, to, uids, "-FLAGS", cmds["remove_flags"].([]interface{}), false); err != nil {
			return err
		}
	}

	if cmds["replace_all_flags"] != nil {
		if err := imap.SetMailFlags(c, to, uids, "FLAGS", cmds["replace_all_flags"].([]interface{}), false); err != nil {
			return err
		}
	}

	return err
}
