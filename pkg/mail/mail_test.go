package mail

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
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

	accs := map[string]*config.Account{
		"main": integration.NewStandardAccount(t),
	}

	defer func() {
		for _, err := range conn.DisconnectAll(accs) {
			assert.Nil(err)
		}
	}()

	// ACTUAL TESTS BELOW

	var err error
	var uids []uint32
	require.Nil(conn.Connect(accs["main"]))

	// Upload test mails first
	for i := 1; i <= 10; i++ {
		assert.Nil(UploadMails(accs["main"], fmt.Sprintf("../../test/data/mails/log%v.txt", i), accs["main"].Connection.InputMailbox.Mailbox, []string{}))
	}

	// Test searching
	uids, err = searchMails(accs["main"])
	assert.Equal([]uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, uids)
	assert.Nil(err)

	// Load newly uploaded mails
	var fetchedMails []*imap.Message
	fetchedMails, err = SearchAndFetchMails(accs["main"])
	assert.Equal(10, len(fetchedMails))
	assert.Nil(err)
}
