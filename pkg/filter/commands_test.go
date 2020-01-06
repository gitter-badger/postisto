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
	const numTestmails = 2

	defer func() {
		require.Nil(conn.Disconnect(acc.Connection.Client))
	}()

	var err error
	acc.Connection.Client, err = conn.Connect(acc.Connection)
	require.Nil(err)

	for i := 1; i <= numTestmails; i++ {
		require.Nil(mail.UploadMails(acc.Connection.Client, fmt.Sprintf("../../test/data/mails/log%v.txt", i), "INBOX", []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	testMails, err := mail.SearchAndFetchMails(acc.Connection.Client, "INBOX", nil, nil)
	require.Equal(numTestmails, len(testMails))
	require.Nil(err)

	// Apply commands
	cmds := make(config.FilterOps)
	cmds["move"] = "MyTarget"
	cmds["add_flags"] = []interface{}{"add_foobar", "Bar", "$MailFlagBit0", imap.FlaggedFlag}
	cmds["remove_flags"] = []interface{}{"set_foobar", "bar"}

	// Mail 1
	require.Nil(RunCommands(acc.Connection.Client, "INBOX", testMails[0].RawMail.Uid, cmds))
	flags, err := mail.GetMailFlags(acc.Connection.Client, "MyTarget", testMails[0].RawMail.Uid)
	require.Nil(err)
	require.ElementsMatch([]string{"add_foobar", "$mailflagbit0", imap.FlaggedFlag}, flags)

	// Mail 2: replace all flags
	cmds["replace_all_flags"] = []interface{}{"42", "bar", "oO", "$MailFlagBit0", imap.FlaggedFlag}
	require.Nil(RunCommands(acc.Connection.Client, "INBOX", testMails[1].RawMail.Uid, cmds))
	flags, err = mail.GetMailFlags(acc.Connection.Client, "MyTarget", testMails[1].RawMail.Uid)
	require.Nil(err)
	require.ElementsMatch([]string{"42", "bar", "oo", "$mailflagbit0", imap.FlaggedFlag}, flags)

	// Upload fresh mail
	require.Nil(mail.UploadMails(acc.Connection.Client, fmt.Sprintf("../../test/data/mails/log%v.txt", 1), "INBOX", []string{}))

	// Load newly uploaded mail
	testMails, err = mail.SearchAndFetchMails(acc.Connection.Client, "INBOX", nil, nil)
	require.Equal(1, len(testMails))
	require.Nil(err)

	// Apply cmd to this new mail 3 too
	cmds["replace_all_flags"] = []interface{}{"completly", "different"}
	require.Nil(RunCommands(acc.Connection.Client, "INBOX", testMails[0].RawMail.Uid, cmds))
	flags, err = mail.GetMailFlags(acc.Connection.Client, "MyTarget", testMails[0].RawMail.Uid)
	require.Nil(err)
	require.ElementsMatch([]string{"completly", "different"}, flags)

	// Verify resulting INBOX
	uids, err := mail.SearchMails(acc.Connection.Client, "INBOX", nil, nil)
	require.Nil(err)
	require.Empty(uids)

	// Verify resulting MyTarget
	uids, err = mail.SearchMails(acc.Connection.Client, "MyTarget", nil, nil)
	require.Nil(err)
	require.ElementsMatch([]uint32{1, 2, 3}, uids)
}
