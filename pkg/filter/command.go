package filter

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/mail"
)

type UnknownCommandTypeError struct {
	typeName string
}

func (err *UnknownCommandTypeError) Error() string {
	return fmt.Sprintf("Command type %q unknown", err.typeName)
}

type BadCommandTargetError struct {
	targetName string
}

func (err *BadCommandTargetError) Error() string {
	return fmt.Sprintf("Bad command target %q", err.targetName)
}

func ApplyCommand(acc *config.Account, mailbox string, uid uint32, cmds []config.Command) error {
	var err error

	for _, cmd := range cmds {

		switch cmd.Type {
		case "move":
			if cmd.Target == "" {
				return &BadCommandTargetError{targetName: cmd.Target}
			}

			if err := mail.MoveMail(acc, uid, cmd.Target); err != nil {
				return err
			}

			if len(cmd.AddFlags) > 0 {
				if err := mail.SetMailFlags(acc, mailbox, uid, "+FLAGS", cmd.AddFlags); err != nil {
					return err
				}
			}

		default:
			return &UnknownCommandTypeError{typeName: cmd.Type}
		}
		//acc.Connection.Client.UidCopy(0, )

		//fmt.Println(mail, filter)
	}
	return err
}
