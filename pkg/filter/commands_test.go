package filter_test

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/filter"
	"github.com/arnisoph/postisto/pkg/imap"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestApplyCommands(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 2

	var err error
	imapClient, err := imap.NewClient(acc.Connection)
	require.Nil(err)

	defer func() {
		require.Nil(imapClient.Disconnect())
	}()

	for i := 1; i <= numTestmails; i++ {
		require.Nil(imapClient.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", i), "INBOX", []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	testMails, err := imapClient.SearchAndFetch("INBOX", nil, nil)
	require.Equal(numTestmails, len(testMails))
	require.Nil(err)

	// Apply commands
	cmds := make(config.FilterOps)
	cmds["move"] = "MyTarget"
	cmds["add_flags"] = []interface{}{"add_foobar", "Bar", "$MailFlagBit0", imap.FlaggedFlag}
	cmds["remove_flags"] = []interface{}{"set_foobar", "bar"}

	// Message 1
	require.Nil(filter.RunCommands(imapClient, "INBOX", testMails[0].RawMessage.Uid, cmds))
	flags, err := imapClient.GetFlags("MyTarget", testMails[0].RawMessage.Uid)
	require.Nil(err)
	require.ElementsMatch([]string{"add_foobar", "$mailflagbit0", imap.FlaggedFlag}, flags)

	// Message 2: replace all flags
	cmds["replace_all_flags"] = []interface{}{"42", "bar", "oO", "$MailFlagBit0", imap.FlaggedFlag}
	require.Nil(filter.RunCommands(imapClient, "INBOX", testMails[1].RawMessage.Uid, cmds))
	flags, err = imapClient.GetFlags("MyTarget", testMails[1].RawMessage.Uid)
	require.Nil(err)
	require.ElementsMatch([]string{"42", "bar", "oo", "$mailflagbit0", imap.FlaggedFlag}, flags)

	// Upload fresh mail
	require.Nil(imapClient.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", 1), "INBOX", []string{}))

	// Load newly uploaded mail
	testMails, err = imapClient.SearchAndFetch("INBOX", nil, nil)
	require.Equal(1, len(testMails))
	require.Nil(err)

	// Apply cmd to this new mail 3 too
	cmds["replace_all_flags"] = []interface{}{"completly", "different"}
	require.Nil(filter.RunCommands(imapClient, "INBOX", testMails[0].RawMessage.Uid, cmds))
	flags, err = imapClient.GetFlags("MyTarget", testMails[0].RawMessage.Uid)
	require.Nil(err)
	require.ElementsMatch([]string{"completly", "different"}, flags)

	// Verify resulting INBOX
	uids, err := imapClient.Search("INBOX", nil, nil)
	require.Nil(err)
	require.Empty(uids)

	// Verify resulting MyTarget
	uids, err = imapClient.Search("MyTarget", nil, nil)
	require.Nil(err)
	require.ElementsMatch([]uint32{1, 2, 3}, uids)
}
