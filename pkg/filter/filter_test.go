package filter_test

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/filter"
	"github.com/arnisoph/postisto/pkg/imap"
	"github.com/arnisoph/postisto/pkg/log"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/stretchr/testify/require"
	"os"
	"strings"
	"testing"
)

func TestGetUnsortedMails(t *testing.T) {
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
	testMessages, err := filter.GetUnsortedMsgs(imapClient, *acc.InputMailbox)
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
		fallbackMsgNum int
		mailsToUpload  []int
		targets        []targetStruct
	}
	tests := []parserTest{
		{ // #1
			fallbackMsgNum: 1,
			mailsToUpload:  []int{1, 2, 3, 4},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
			},
		},
		{ // #2
			mailsToUpload: []int{1, 2, 3, 4},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-TestAND", num: 1},
			},
		},
		{ // #3
			mailsToUpload: []int{1, 2, 3, 4},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-TestRegex", num: 1},
			},
		},
		{ // #4
			mailsToUpload: []int{1, 2, 3, 8, 9, 14},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-TestMisc", num: 3},
			},
		},
		{ // #5
			mailsToUpload: []int{1, 2, 3, 13, 15, 16, 17},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-TestUnicodeFrom-梦龙周", num: 3},
				{name: "MailFilterTest-TestUnicodeSubject", num: 1},
			},
		},
		{ // #6
			mailsToUpload: []int{1, 2, 3, 16, 17},
			targets: []targetStruct{
				{name: "MyTarget", num: 3},
				{name: "MailFilterTest-foo", num: 1},
				{name: "MailFilterTest-bar", num: 1},
			},
		},
		{ // #7
			fallbackMsgNum: 3,
			mailsToUpload:  []int{1, 2, 3, 16, 17},
			targets: []targetStruct{
				{name: "MailFilterTest-baz", num: 1},
				{name: "MailFilterTest-zab", num: 1},
			},
		},
		{ // #8
			fallbackMsgNum: 3,
			mailsToUpload:  []int{1, 2, 3, 16, 17},
			targets: []targetStruct{
				{name: "X-Postisto-MailFilterTest-lorem", num: 1},
				{name: "X-Postisto-MailFilterTest-ipsum", num: 1},
			},
		},
	}

	for testNum, test := range tests {
		log.Debug(fmt.Sprintf("Starting TestEvaluateFilterSetsOnMails #%v", testNum+1))

		// Get config
		cfg, err := config.NewConfigFromFile(fmt.Sprintf("../../test/data/configs/valid/local_imap_server/TestEvaluateFilterSetsOnMails-%v/", testNum+1))
		require.Nil(err)
		acc := cfg.Accounts["local_imap_server"]

		if acc.Connection.Server == "" {
			// Create new random user
			acc.Connection = integration.NewStandardAccount(t).Connection
		}

		if strings.Contains(acc.Connection.Server, "gmail") {
			acc.Connection.Username = os.Getenv("POSTISTO_GMAIL_TEST_ACC_USERNAME")
			acc.Connection.Password = os.Getenv("POSTISTO_GMAIL_TEST_ACC_PASSWORD")
		}

		// Set debug info for failed assertions
		debugInfo := map[string]string{"username": acc.Connection.Username, "testNum": fmt.Sprint(testNum + 1)}

		// NewClient to IMAP server
		imapClient, err := imap.NewClient(acc.Connection)
		require.Nil(err)

		if strings.Contains(acc.Connection.Server, "gmail") {
			log.Debug("Detected gmail account. Going to cleanup...")
			uids, err := imapClient.Search("INBOX", nil, nil)
			require.NoError(err)
			if len(uids) > 0 {
				err = imapClient.DeleteMsgs("INBOX", uids, true)
				require.NoError(err)
			}

			mailBoxes, err := imapClient.List()
			require.NoError(err)
			for mailboxName, _ := range mailBoxes {
				if strings.Contains(strings.ToLower(mailboxName), "x-postisto") {
					require.NoError(imapClient.DeleteMailbox(mailboxName))
				}
			}
		}

		// Simulate new unsorted mails by uploading
		for i, mailNum := range test.mailsToUpload {
			require.NotNil(imapClient)
			require.NotNil(acc)
			require.NotNil(acc.InputMailbox)
			require.Nil(imapClient.Upload(fmt.Sprintf("../../test/data/mails/log%v.txt", mailNum), acc.InputMailbox.Mailbox, []string{}), debugInfo)

			var withoutFlags []string
			if !strings.Contains(acc.Connection.Server, "gmail") { // gmail does some extra magic, marking (some) new messages as "important"....
				withoutFlags = append(withoutFlags,imap.FlaggedFlag)
			}

			// verify upload
			uploadedMails, err :=imapClient.Search(acc.InputMailbox.Mailbox, nil, withoutFlags)

			require.NoError(err)
			require.Len(uploadedMails, i+1, fmt.Sprintf("This (#%v) or one of the previous mail uploads failed!",i+1),debugInfo)
		}

		// ACTUAL TESTS BELOW

		// Baaaam
		_, err = filter.EvaluateFilterSetsOnMsgs(imapClient, *acc)
		require.Nil(err, debugInfo)


		fallbackMethod := "moving"
		if *acc.FallbackMailbox == acc.InputMailbox.Mailbox || *acc.FallbackMailbox == "" {
			fallbackMethod = "flagging"
		}

		// Verify Source
		if fallbackMethod == "flagging" {
				fetchedMails, err := imapClient.Search(acc.InputMailbox.Mailbox, nil, []string{imap.FlaggedFlag})
				require.Nil(err, debugInfo)
				require.Equal(0, len(fetchedMails), "Unexpected num of mails in source %v", acc.InputMailbox.Mailbox, debugInfo)
		} else {
			// fallback = moving
			fetchedMails, err := imapClient.Search(acc.InputMailbox.Mailbox, nil, nil)
			require.Nil(err, debugInfo)
			require.Equal(0, len(fetchedMails), "Unexpected num of mails in source %v", acc.InputMailbox.Mailbox, debugInfo)
		}

		// Verify Targets
		for _, target := range test.targets {
			// fallback = flagging
			fetchedMails, err := imapClient.Search(target.name, nil, nil)
			require.Nil(err, debugInfo)
			require.Equal(target.num, len(fetchedMails), "Unexpected num of mails in target %v", target.name, debugInfo)
		}

		// Verify fallback mailbox (if != source)
		if acc.InputMailbox.Mailbox != *acc.FallbackMailbox {
			fallBackMsgs, err := imapClient.Search(*acc.FallbackMailbox, nil, nil)
			require.Nil(err, debugInfo)
			require.Equal(test.fallbackMsgNum, len(fallBackMsgs), debugInfo)
		}

		// Disconnect - Hoooraaay!
		require.Nil(imapClient.Disconnect())
	}
}
