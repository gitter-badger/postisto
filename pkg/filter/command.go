package filter

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/mail"
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

func ApplyCommands(acc *config.Account, from string, to string, uid uint32, cmds config.Commands) error {
	var err error

	if cmds["move"] != nil {
		if err := mail.MoveMail(acc, uid, from, cmds["move"].(string)); err != nil {
			return err
		}
	}

	if cmds["add_flags"] != nil {
		if err := mail.SetMailFlags(acc, to, uid, "+FLAGS", cmds["add_flags"].([]interface{})); err != nil {
			return err
		}
	}

	if cmds["remove_flags"] != nil {
		if err := mail.SetMailFlags(acc, to, uid, "-FLAGS", cmds["remove_flags"].([]interface{})); err != nil {
			return err
		}
	}


	if cmds["replace_all_flags"] != nil {
		if err := mail.SetMailFlags(acc, to, uid, "FLAGS", cmds["replace_all_flags"].([]interface{})); err != nil {
			return err
		}
	}

	return err
}
