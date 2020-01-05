package filter

import (
	"fmt"
	"github.com/emersion/go-imap"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/conn"
	"github.com/arnisoph/postisto/pkg/mail"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestApplyCommands(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 8

	defer func() {
		require.Nil(conn.Disconnect(acc.Connection.Client))
	}()

	var err error
	acc.Connection.Client, err = conn.Connect(acc.Connection)
	require.Nil(err)

	for i := 1; i <= numTestmails; i++ {
		require.Nil(mail.UploadMails(acc.Connection.Client, fmt.Sprintf("../../test/data/mails/log%v.txt", i), acc.Connection.InputMailbox.Mailbox, []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	fetchedMails, err := mail.SearchAndFetchMails(acc.Connection.Client, acc.Connection.InputMailbox.Mailbox, nil, nil)
	require.Equal(numTestmails, len(fetchedMails))
	require.Nil(err)

	// Apply commands
	cmds := make(config.Commands)
	cmds["move"] = "MyTarget"
	//cmds["replace_all_flags"] = []interface{}{"42", "bar", "oO", "$MailFlagBit0", imap.FlaggedFlag}
	cmds["add_flags"] = []interface{}{"add_foobar", "Bar", "$MailFlagBit0", imap.FlaggedFlag}
	cmds["remove_flags"] = []interface{}{"set_foobar", "bar"}

	var uids []uint32
	for _, fetchedMail := range fetchedMails {
		uids = append(uids, fetchedMail.Uid)
		require.Nil(RunCommands(acc.Connection.Client, acc.Connection.InputMailbox.Mailbox, "MyTarget", fetchedMail.Uid, cmds))

		flags, err := mail.GetMailFlags(acc.Connection.Client, "MyTarget", fetchedMail.Uid)
		require.Nil(err)
		require.ElementsMatch([]interface{}{"add_foobar", "$mailflagbit0", imap.FlaggedFlag}, flags)
	}

	movedMails, err := mail.FetchMails(acc.Connection.Client, "MyTarget", uids)
	require.Nil(err)
	require.EqualValues(numTestmails, len(movedMails))

	//oldMails, err := mail.Sea

}
