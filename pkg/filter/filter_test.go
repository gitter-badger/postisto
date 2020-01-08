package filter

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/conn"
	"github.com/arnisoph/postisto/pkg/mail"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetUnsortedMails(t *testing.T) {
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
	testMessages, err := getUnsortedMails(acc.Connection.Client, *acc.InputMailbox)
	require.Nil(err)
	require.Equal(2, len(testMessages))
}

func TestEvaluateFilterSetsOnMails(t *testing.T) {
	require := require.New(t)

	type targetStruct struct {
		name string
		num  int
	}
	type parserTest struct {
		source          string
		sourceRemaining int
		inboxNum        int
		mailsToUpload   []int
		targets         []targetStruct
	}
	tests := []parserTest{
		{ // #1
			source:          "INBOX",
			sourceRemaining: 1,
			mailsToUpload:   []int{1, 2, 3, 4},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
			},
		},
		{ // #2
			source:        "INBOX",
			mailsToUpload: []int{1, 2, 3, 4},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-TestAND", num: 1},
			},
		},
		{ // #3
			source:        "INBOX",
			mailsToUpload: []int{1, 2, 3, 4},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-TestRegex", num: 1},
			},
		},
		{ // #4
			source:        "INBOX",
			mailsToUpload: []int{1, 2, 3, 8, 9, 14},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-TestMisc", num: 3},
			},
		},
		{ // #5
			source:        "INBOX",
			mailsToUpload: []int{1, 2, 3, 13, 15, 16, 17},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-TestUnicodeFrom-梦龙周", num: 3},
				{name: "MailFilterTest-TestUnicodeSubject", num: 1},
			},
		},
		{ // #6
			source:        "INBOX",
			mailsToUpload: []int{1, 2, 3, 16, 17},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-16", num: 1},
				{name: "MailFilterTest-17", num: 1},
			},
		},
		{ // #7
			source:        "PreInbox",
			inboxNum:      3,
			mailsToUpload: []int{1, 2, 3, 16, 17},
			targets: []targetStruct{
				{name: "MailFilterTest-16", num: 1},
				{name: "MailFilterTest-17", num: 1},
			},
		},
	}

	for testNum, test := range tests {
		// Get config
		cfg, err := config.NewConfigFromFile(fmt.Sprintf("../../test/data/configs/valid/local_imap_server/TestEvaluateFilterSetsOnMails-%v/", testNum+1))
		require.Nil(err)
		acc := cfg.Accounts["local_imap_server"]

		// Create new random user
		acc.Connection = integration.NewStandardAccount(t).Connection

		// Set debug info for failed assertions
		debugInfo := map[string]string{"username": acc.Connection.Username, "testNum": fmt.Sprint(testNum + 1)}

		// Connect to IMAP server
		acc.Connection.Client, err = conn.Connect(acc.Connection)
		require.Nil(err)

		// Simulate new unsorted mails by uploading
		for _, mailNum := range test.mailsToUpload {
			require.Nil(mail.UploadMails(acc.Connection.Client, fmt.Sprintf("../../test/data/mails/log%v.txt", mailNum), acc.InputMailbox.Mailbox, []string{}), debugInfo)
		}

		// ACTUAL TESTS BELOW

		// Baaaam
		_, err = EvaluateFilterSetsOnMails(*acc)
		require.Nil(err)

		// Verify Source
		fetchedMails, err := mail.SearchMails(acc.Connection.Client, acc.InputMailbox.Mailbox, nil, nil)
		require.Nil(err, debugInfo)
		//require.Equal(test.sourceNum, len(remainingMails))
		require.Equal(test.sourceRemaining, len(fetchedMails), "Unexpected num of mails in source %v", acc.InputMailbox.Mailbox, debugInfo)

		// Verify Targets
		for _, target := range test.targets {
			fetchedMails, err := mail.SearchMails(acc.Connection.Client, target.name, nil, nil)
			require.Nil(err, debugInfo)
			require.Equal(target.num, len(fetchedMails), "Unexpected num of mails in target %v", target.name, debugInfo)
		}

		// Verify fallback mailbox (if != source)
		if acc.InputMailbox.Mailbox != *acc.FallbackMailbox {
			inboxMails, err := mail.SearchMails(acc.Connection.Client, *acc.FallbackMailbox, nil, nil)
			require.Nil(err)
			require.Equal(test.inboxNum, len(inboxMails))
		}

		// Disconnect - Hoooraaay!
		require.Nil(conn.Disconnect(acc.Connection.Client))
	}
}
