package mail

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/conn"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/emersion/go-imap"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestUploadMails(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)

	defer func() {
		require.Nil(conn.Disconnect(acc.Connection.Client))
	}()

	var err error
	acc.Connection.Client, err = conn.Connect(acc.Connection)
	require.Nil(err)

	require.EqualError(UploadMails(acc.Connection.Client, "does-not-exit.txt", "INBOX", []string{}), "open does-not-exit.txt: no such file or directory")
	require.Error(UploadMails(acc.Connection.Client, "../../test/data/mails/empty-mail.txt", "INBOX", []string{}))
}

func TestSearchAndFetchMails(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 3

	defer func() {
		require.Nil(conn.Disconnect(acc.Connection.Client))
	}()

	var err error
	acc.Connection.Client, err = conn.Connect(acc.Connection)
	require.Nil(err)

	for i := 1; i <= numTestmails; i++ {
		require.Nil(UploadMails(acc.Connection.Client, fmt.Sprintf("../../test/data/mails/log%v.txt", i), "INBOX", []string{}))
	}

	// ACTUAL TESTS BELOW

	// Test searching only
	uids, err := SearchMails(acc.Connection.Client, "INBOX", nil, nil)
	require.Nil(err)
	require.Equal([]uint32{1, 2, 3}, uids)

	// Load newly uploaded mails
	var fetchedMails []*imap.Message

	// Search in non-existing mailbox
	fetchedMails, err = SearchAndFetchMails(acc.Connection.Client, "non-existent", nil, nil)
	require.True(strings.HasPrefix(err.Error(), "Mailbox doesn't exist: non-existent"))
	require.Equal(0, len(fetchedMails))

	fetchedMails, err = SearchAndFetchMails(acc.Connection.Client, "INBOX", nil, nil)
	require.Nil(err)
	require.Equal(numTestmails, len(fetchedMails))
}

func TestSetMailFlags(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 6

	defer func() {
		require.Nil(conn.Disconnect(acc.Connection.Client))
	}()

	var err error
	acc.Connection.Client, err = conn.Connect(acc.Connection)
	require.Nil(err)

	for i := 1; i <= numTestmails; i++ {
		require.Nil(UploadMails(acc.Connection.Client, fmt.Sprintf("../../test/data/mails/log%v.txt", i), "INBOX", []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	fetchedMails, err := SearchAndFetchMails(acc.Connection.Client, "INBOX", nil, nil)
	require.Nil(err)
	require.Equal(numTestmails, len(fetchedMails))

	// Set custom flags
	for _, fetchedMail := range fetchedMails {
		var flags []string

		// Add flags
		require.Nil(SetMailFlags(acc.Connection.Client, "INBOX", []uint32{fetchedMail.Uid}, "+FLAGS", []interface{}{"fooooooo", "asdasd", "$MailFlagBit0", imap.FlaggedFlag}))
		flags, err = GetMailFlags(acc.Connection.Client, "INBOX", fetchedMail.Uid)
		require.Nil(err)
		require.ElementsMatch([]string{"fooooooo", "asdasd", "$mailflagbit0", imap.FlaggedFlag}, flags)

		// Remove flags
		require.Nil(SetMailFlags(acc.Connection.Client, "INBOX", []uint32{fetchedMail.Uid}, "-FLAGS", []interface{}{"fooooooo", "asdasd"}))
		flags, err = GetMailFlags(acc.Connection.Client, "INBOX", fetchedMail.Uid)
		require.Nil(err)
		require.ElementsMatch([]string{"$mailflagbit0", imap.FlaggedFlag}, flags)

		// Replace all flags with new list
		require.Nil(SetMailFlags(acc.Connection.Client, "INBOX", []uint32{fetchedMail.Uid}, "FLAGS", []interface{}{"123", "forty-two"}))
		flags, err = GetMailFlags(acc.Connection.Client, "INBOX", fetchedMail.Uid)
		require.Nil(err)
		require.ElementsMatch([]string{"123", "forty-two"}, flags)
	}
}
