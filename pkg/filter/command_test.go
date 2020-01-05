package filter

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/conn"
	"github.com/arnisoph/postisto/pkg/mail"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/emersion/go-imap"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestApplyCommands(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)

	defer func() {
		require.Nil(conn.Disconnect(acc))
	}()

	require.Nil(conn.Connect(acc))

	for i := 1; i <= 3; i++ {
		require.Nil(mail.UploadMails(acc, fmt.Sprintf("../../test/data/mails/log%v.txt", i), acc.Connection.InputMailbox.Mailbox, []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	fetchedMails, err := mail.SearchAndFetchMails(acc)
	require.Equal(3, len(fetchedMails))
	require.Nil(err)

	// Apply commands
	cmds := make(config.Commands)
	cmds["move"] = "MyTarget"
	//cmds["replace_all_flags"] = []interface{}{"42", "bar", "oO", "$MailFlagBit0", imap.FlaggedFlag}
	cmds["add_flags"] = []interface{}{"add_foobar", "Bar", "$MailFlagBit0", imap.FlaggedFlag}
	cmds["remove_flags"] = []interface{}{"set_foobar", "bar"}

	for _, fetchedMail := range fetchedMails {
		require.Nil(ApplyCommands(acc, acc.Connection.InputMailbox.Mailbox, "MyTarget", fetchedMail.Uid, cmds))
		flags, err := mail.GetMailFlags(acc, "MyTarget", fetchedMail.Uid)
		require.Nil(err)
		require.ElementsMatch([]interface{}{"add_foobar", "$mailflagbit0", imap.FlaggedFlag}, flags)
	}

}
