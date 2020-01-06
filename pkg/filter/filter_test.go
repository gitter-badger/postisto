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
	testMessages, err := getUnsortedMails(acc.Connection.Client, *acc.Connection.InputMailbox)
	require.Nil(err)
	require.Equal(2, len(testMessages))
}

func TestEvaluateFilterSetOnMails(t *testing.T) {
	require := require.New(t)

	// Get config
	cfg := config.New()
	cfg, err := cfg.Load("../../test/data/configs/valid/")
	require.Nil(err)
	acc := cfg.Accounts["local_imap_server"]

	// Create new random user
	acc.Connection = integration.NewStandardAccount(t).Connection

	// Connect to IMAP server
	acc.Connection.Client, err = conn.Connect(acc.Connection)
	require.Nil(err)

	defer func() {
		require.Nil(conn.Disconnect(acc.Connection.Client))
	}()

	// Simulate new unsorted mails by uploading
	const numTestmails = 3
	for i := 1; i <= numTestmails; i++ {
		require.Nil(mail.UploadMails(acc.Connection.Client, fmt.Sprintf("../../test/data/mails/log%v.txt", i), "INBOX", []string{}))
	}

	// ACTUAL TESTS BELOW
	matched, err := EvaluateFilterSetOnMails(*acc)
	require.True(matched)
	require.Nil(err)

	//require.Fail("oOoOo")
}
