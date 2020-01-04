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
	fetchedMails, err = SearchAndFetchMails(acc)
	assert.Equal(10, len(fetchedMails))
	assert.Nil(err)
}
