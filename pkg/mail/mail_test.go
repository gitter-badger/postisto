package mail

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/conn"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/emersion/go-imap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUploadMails(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	acc := integration.NewStandardAccount(t)

	defer func() {
		assert.Nil(conn.Disconnect(acc))
	}()

	require.Nil(conn.Connect(acc))

	assert.EqualError(UploadMails(acc, "does-not-exit.txt", acc.Connection.InputMailbox.Mailbox, []string{}), "open does-not-exit.txt: no such file or directory")
	assert.Error(UploadMails(acc, "../../test/data/mails/empty-mail.txt", acc.Connection.InputMailbox.Mailbox, []string{}))
}

func TestSearchAndFetchMails(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	acc := integration.NewStandardAccount(t)

	defer func() {
		assert.Nil(conn.Disconnect(acc))
	}()

	require.Nil(conn.Connect(acc))

	for i := 1; i <= 10; i++ {
		require.Nil(UploadMails(acc, fmt.Sprintf("../../test/data/mails/log%v.txt", i), acc.Connection.InputMailbox.Mailbox, []string{}))
	}

	// ACTUAL TESTS BELOW

	// Test searching only
	uids, err := searchMails(acc)
	assert.Equal([]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, uids)
	assert.Nil(err)

	// Load newly uploaded mails
	var fetchedMails []*imap.Message

	acc.Connection.InputMailbox.Mailbox = "wrong"
	fetchedMails, err = SearchAndFetchMails(acc)
	assert.Equal(0, len(fetchedMails))
	assert.Nil(err)
	acc.Connection.InputMailbox.Mailbox = "INBOX"

	fetchedMails, err = SearchAndFetchMails(acc)
	assert.Equal(10, len(fetchedMails))
	assert.Nil(err)
}

func TestSetMailFlags(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	acc := integration.NewStandardAccount(t)

	defer func() {
		assert.Nil(conn.Disconnect(acc))
	}()

	require.Nil(conn.Connect(acc))

	for i := 1; i <= 10; i++ {
		require.Nil(UploadMails(acc, fmt.Sprintf("../../test/data/mails/log%v.txt", i), acc.Connection.InputMailbox.Mailbox, []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	fetchedMails, err := SearchAndFetchMails(acc)
	assert.Equal(10, len(fetchedMails))
	assert.Nil(err)

	// Set custom flags
	for _, fetchedMail := range fetchedMails {
		assert.Nil(SetMailFlags(acc, acc.Connection.InputMailbox.Mailbox, fetchedMail.Uid, "+FLAGS", []interface{}{"fooooooo", "asdasd", "$MailFlagBit0", imap.FlaggedFlag}))

		flags, err := GetMailFlags(acc, acc.Connection.InputMailbox.Mailbox, fetchedMail.Uid)
		assert.Nil(err)
		assert.ElementsMatch([]string{"fooooooo", "asdasd", "$mailflagbit0", imap.FlaggedFlag}, flags)
	}
}
