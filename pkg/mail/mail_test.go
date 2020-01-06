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

	// ACTUAL TESTS BELOW

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

	// Search in non-existing mailbox
	fetchedMails, err := SearchAndFetchMails(acc.Connection.Client, "non-existent", nil, nil)
	require.Error(err)
	require.True(strings.HasPrefix(err.Error(), "Mailbox doesn't exist: non-existent"))
	require.Equal(0, len(fetchedMails))

	// Search in correct Mailbox now
	fetchedMails, err = SearchAndFetchMails(acc.Connection.Client, "INBOX", nil, nil)
	require.Nil(err)
	require.Equal(numTestmails, len(fetchedMails))
}

func TestSetMailFlags(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 1

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

	// Test failed GetMailFlags (because of non-existing mailbox)
	_, err = GetMailFlags(acc.Connection.Client, "non-existing-mailbox", 0)
	require.Error(err)
	require.True(strings.HasPrefix(err.Error(), "Mailbox doesn't exist: non-existing-mailbox"))

	// Set custom flags
	var flags []string

	// Add flags
	require.Nil(SetMailFlags(acc.Connection.Client, "INBOX", []uint32{fetchedMails[0].RawMail.Uid}, "+FLAGS", []interface{}{"fooooooo", "asdasd", "$MailFlagBit0", imap.FlaggedFlag}, false))
	flags, err = GetMailFlags(acc.Connection.Client, "INBOX", fetchedMails[0].RawMail.Uid)
	require.Nil(err)
	require.ElementsMatch([]string{"fooooooo", "asdasd", "$mailflagbit0", imap.FlaggedFlag}, flags)

	// Remove flags
	require.Nil(SetMailFlags(acc.Connection.Client, "INBOX", []uint32{fetchedMails[0].RawMail.Uid}, "-FLAGS", []interface{}{"fooooooo", "asdasd"}, false))
	flags, err = GetMailFlags(acc.Connection.Client, "INBOX", fetchedMails[0].RawMail.Uid)
	require.Nil(err)
	require.ElementsMatch([]string{"$mailflagbit0", imap.FlaggedFlag}, flags)

	// Replace all flags with new list
	require.Nil(SetMailFlags(acc.Connection.Client, "INBOX", []uint32{fetchedMails[0].RawMail.Uid}, "FLAGS", []interface{}{"123", "forty-two"}, false))
	flags, err = GetMailFlags(acc.Connection.Client, "INBOX", fetchedMails[0].RawMail.Uid)
	require.Nil(err)
	require.ElementsMatch([]string{"123", "forty-two"}, flags)
}

func TestMoveMails(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 5

	defer func() {
		require.Nil(conn.Disconnect(acc.Connection.Client))
	}()

	var err error
	acc.Connection.Client, err = conn.Connect(acc.Connection)
	require.Nil(err)

	for i := 1; i <= numTestmails; i++ {
		require.Nil(UploadMails(acc.Connection.Client, fmt.Sprintf("../../test/data/mails/log%v.txt", i), acc.InputMailbox.Mailbox, []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	fetchedMails, err := SearchAndFetchMails(acc.Connection.Client, acc.InputMailbox.Mailbox, nil, nil)
	require.Equal(numTestmails, len(fetchedMails))
	require.Nil(err)

	// Move mails arround
	err = MoveMails(acc.Connection.Client, []uint32{fetchedMails[0].RawMail.Uid}, "INBOX", "MyTarget42")
	require.Nil(err)

	err = MoveMails(acc.Connection.Client, []uint32{fetchedMails[1].RawMail.Uid}, "INBOX", "INBOX")
	require.Nil(err)

	err = MoveMails(acc.Connection.Client, []uint32{fetchedMails[2].RawMail.Uid}, "INBOX", "MyTarget!!!")
	require.Nil(err)

	err = MoveMails(acc.Connection.Client, []uint32{fetchedMails[3].RawMail.Uid}, "wrong-source", "MyTarget!!!")
	require.True(strings.HasPrefix(err.Error(), "Mailbox doesn't exist: wrong-source"))

	err = MoveMails(acc.Connection.Client, []uint32{fetchedMails[4].RawMail.Uid}, "INBOX", "ütf-8 & 梦龙周")
	require.Nil(err)

	var uids []uint32
	uids, err = SearchMails(acc.Connection.Client, "INBOX", nil, nil)
	require.Nil(err)
	require.EqualValues([]uint32{4, 6}, uids) // UID 1 moved, UID 2 became 6, UID 3 moved, UID 4 kept untouched, UID 5 moved
}

func TestDeleteMails(t *testing.T) {
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
		require.Nil(UploadMails(acc.Connection.Client, fmt.Sprintf("../../test/data/mails/log%v.txt", i), acc.InputMailbox.Mailbox, []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	fetchedMails, err := SearchAndFetchMails(acc.Connection.Client, acc.InputMailbox.Mailbox, nil, nil)
	require.Equal(numTestmails, len(fetchedMails))
	require.Nil(err)

	// Delete one mail
	err = DeleteMails(acc.Connection.Client, "does-not-exist", []uint32{fetchedMails[0].RawMail.Uid}, true) // mailbox doesn't exist, can't be deleted
	require.True(strings.HasPrefix(err.Error(), "Mailbox doesn't exist: does-not-exist"))

	err = DeleteMails(acc.Connection.Client, "INBOX", []uint32{fetchedMails[1].RawMail.Uid}, false) // not moved yet, flag, don't expunge yet
	require.Nil(err)
	flags, err := GetMailFlags(acc.Connection.Client, "INBOX", fetchedMails[1].RawMail.Uid)
	require.Nil(err)
	require.EqualValues([]string{imap.DeletedFlag}, flags)
	err = DeleteMails(acc.Connection.Client, "INBOX", []uint32{fetchedMails[1].RawMail.Uid}, true) // not moved yet, flag & expunge
	require.Nil(err)

	var uids []uint32
	uids, err = SearchMails(acc.Connection.Client, "INBOX", nil, nil)
	require.Nil(err)
	require.EqualValues([]uint32{1, 3}, uids) // UID 1 kept untouched, UID 2 deleted, UID 3 kept untouched
}

func TestParseMailHeaders(t *testing.T) {
	require := require.New(t)

	acc := integration.NewStandardAccount(t)
	const numTestmails = 5

	defer func() {
		require.Nil(conn.Disconnect(acc.Connection.Client))
	}()

	var err error
	acc.Connection.Client, err = conn.Connect(acc.Connection)
	require.Nil(err)

	for i := 1; i <= numTestmails; i++ {
		require.Nil(UploadMails(acc.Connection.Client, fmt.Sprintf("../../test/data/mails/log%v.txt", i), acc.InputMailbox.Mailbox, []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	fetchedMails, err := SearchAndFetchMails(acc.Connection.Client, acc.InputMailbox.Mailbox, nil, nil)
	require.Nil(err)
	require.Equal(numTestmails, len(fetchedMails))

	// Verify parsed fields (header)
	parserTests := []struct {
		from    string
		to      string
		subject string
		date    string
	}{
		{ // #1
			from:    `youth4work <admin@youth4work.com>`,
			to:      "<shubham@cyberzonec.in>",
			subject: "your registration at www.youth4work.com",
		},
		{ // #2
			from:    `rachit <rachit.jain@youth4work.com>`,
			to:      "<shubham@cyberzonec.in>",
			subject: "cyberzonec",
		},
		{ // #3
			from:    `rachit <rachit.jain@youth4work.com>`,
			to:      "<shubham@cyberzonec.in>",
			subject: "contact ranked talent for hire",
		},
		{ // #4
			from:    `bigrock.com sales team <automail@bigrock.com>`,
			to:      "<shubham@cyberzonec.in>",
			subject: "customer sign up",
		},
		{ // #5
			from:    `invalid-address`,
			to:      `"mr. ütf-8" <foo@bar.net>`,
			subject: "ütf-8 💩",
		},
	}

	// Boring standard headers
	for i := 0; i < numTestmails; i++ {
		require.Equal(parserTests[i].from, fetchedMails[i].Headers["from"], "Failed in test #%v (FROM)", i+1)
		require.Equal(parserTests[i].to, fetchedMails[i].Headers["to"], "Failed in test #%v (TO)", i+1)
		require.Equal(parserTests[i].subject, fetchedMails[i].Headers["subject"], "Failed in test #%v (SUBJECT)", i+1)
	}

	// Exciting custom headers
	require.Equal("<0000014e2ed21bf8-035d1578-a0ac-4afe-a3cb-ea4e65b92143-000000@email.amazonses.com>", fetchedMails[4].Headers["message-id"])
	require.Equal("<http://emailsparrow.com/unsubscribe.php?m=2392014&c=6508a8072980153786bbab4969679c2a&l=19&n=57154>", fetchedMails[4].Headers["list-unsubscribe"])
	require.Equal("2392014", fetchedMails[4].Headers["x-mailer-recptid"])
	require.Equal("foo <a@b.c>, baz <d@e.f>", fetchedMails[4].Headers["cc"])
}
