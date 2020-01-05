package filter

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/arnisoph/postisto/pkg/conn"
	"github.com/arnisoph/postisto/pkg/mail"
	"github.com/arnisoph/postisto/test/integration"
	"github.com/emersion/go-imap"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestApplyCommand(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	acc := integration.NewStandardAccount(t)

	defer func() {
		assert.Nil(conn.Disconnect(acc))
	}()

	require.Nil(conn.Connect(acc))

	for i := 1; i <= 10; i++ {
		require.Nil(mail.UploadMails(acc, fmt.Sprintf("../../test/data/mails/log%v.txt", i), acc.Connection.InputMailbox.Mailbox, []string{}))
	}

	// ACTUAL TESTS BELOW

	// Load newly uploaded mails
	fetchedMails, err := mail.SearchAndFetchMails(acc)
	assert.Equal(10, len(fetchedMails))
	assert.Nil(err)

	// Apply commands
	cmd := []config.Command{}
	cmd = append(cmd, config.Command{Type: "move", Target: "MyTarget", AddFlags: []interface{}{"foobar", "bar", "$MailFlagBit0", imap.FlaggedFlag}})
	for _, fetchedMail := range fetchedMails {
		assert.Nil(ApplyCommand(acc, acc.Connection.InputMailbox.Mailbox, "MyTarget", fetchedMail.Uid, cmd))
		flags, err := mail.GetMailFlags(acc, "MyTarget", fetchedMail.Uid)
		assert.Nil(err)
		assert.ElementsMatch([]string{"foobar", "bar", "$mailflagbit0", imap.FlaggedFlag}, flags)
	}

}
