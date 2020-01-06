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

	type parserTest struct {
		source    string
		sourceNum int
		target    string
		targetNum int
	}
	tests := []parserTest{
		{
			source:    "INBOX",
			sourceNum: 14,
			target:    "MyTarget",
			targetNum: 3,
		},
	}

	for testNum, test := range tests {
		// Get config
		cfg := config.NewConfig()
		cfg, err := cfg.Load(fmt.Sprintf("../../test/data/configs/valid/local_imap_server/TestEvaluateFilterSetsOnMails-%v/", testNum+1))
		require.Nil(err)
		acc := cfg.Accounts["local_imap_server"]

		// Create new random user
		acc.Connection = integration.NewStandardAccount(t).Connection

		// Connect to IMAP server
		acc.Connection.Client, err = conn.Connect(acc.Connection)
		require.Nil(err)

		// Simulate new unsorted mails by uploading
		for mailNum := 1; mailNum <= integration.MaxTestMailCount; mailNum++ {
			require.Nil(mail.UploadMails(acc.Connection.Client, fmt.Sprintf("../../test/data/mails/log%v.txt", mailNum), "INBOX", []string{}))
		}

		_, err = EvaluateFilterSetsOnMails(*acc)
		//require.True(matched)
		require.Nil(err)

		// Verify Source
		fetchedMails, err := mail.SearchMails(acc.Connection.Client, test.source, nil, nil)
		require.Nil(err)
		require.Equal(test.sourceNum, len(fetchedMails))

		// Verify Target
		fetchedMails, err = mail.SearchMails(acc.Connection.Client, test.target, nil, nil)
		require.Nil(err)
		require.Equal(test.targetNum, len(fetchedMails))

		// Disconnect - Hoooraaay!
		require.Nil(conn.Disconnect(acc.Connection.Client))
		break
	}

	//require.Fail("oOoOo")
}
