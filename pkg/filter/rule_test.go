package filter

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestParseRuleSet(t *testing.T) {
	require := require.New(t)

	cfg := config.New()
	cfg, err := cfg.Load("../../test/data/configs/valid/")
	require.Nil(err)

	// ACTUAL TESTS BELOW
	acc := cfg.Accounts["local_imap_server"]

	err = ParseRuleSet(acc.Filters["simple"].RuleSet)
	require.Nil(err)

	//cc.Filters["simple"].RuleSet[0]
}

func TestParseRule(t *testing.T) {
	require := require.New(t)

	// ACTUAL TESTS BELOW

	testMailHeaders := MailHeaders{"from": "foo@example.com", "to": "me@example.com", "subject": "with löve"}

	ruleParserTests := []struct {
		rule          config.Rule
		matchExpected bool
		err           string
	}{
		{ // #1
			rule:          config.Rule{"or": []map[string]interface{}{{"from": "foo@example.com"}}},
			matchExpected: true,
		},
		{ // #2
			rule: config.Rule{
				"or": []map[string]interface{}{
					{"from": "oO"},
					{"from": "foo@example.com"},
				}},
			matchExpected: true,
		},
		{ // #3
			rule:          config.Rule{"or": []map[string]interface{}{{"from": "wrong value"}}},
			matchExpected: false,
		},
		//{
		//	headers: MailHeaders{"from": "oO"},
		//	rule: config.Rule{
		//		"or": []map[string]interface{}{
		//			{
		//				"or": config.Rule{
		//					"or": []map[string]interface{}{
		//						{"from": "nope"},
		//						{"from": "oO"},
		//					},
		//				},
		//			},
		//		},
		//	},
		//	matchExpected: true,
		//},
		{ // #4
			rule: config.Rule{"and": []map[string]interface{}{
				{"from": "foo@example.com"},
				{"to": "me@EXAMPLE.com"},
			}},
			matchExpected: true,
		},
		{ // #5
			rule: config.Rule{"and": []map[string]interface{}{
				{"from": "you"},
				{"to": "you"},
			}},
			matchExpected: false,
		},
		{ // #6
			rule: config.Rule{"and": []map[string]interface{}{
				{"from": "you"},
				{"to": "you"},
			}},
			matchExpected: false,
		},
		{ // #7
			rule: config.Rule{"non-existent-op": []map[string]interface{}{
				{"from": "you"},
				{"to": "you"},
			}},
			matchExpected: false,
			err:           `Rule operator "non-existent-op" is unsupported`,
		},
		{ // #8
			rule:          config.Rule{"and": []map[string]interface{}{{"from": "@example.com"}}},
			matchExpected: true,
		},
		{ // #9
			rule:          config.Rule{"and": []map[string]interface{}{{"from": "@example.com"}}},
			matchExpected: true,
		},
		{ // #10
			rule:          config.Rule{"and": []map[string]interface{}{{"from": ""}}},
			matchExpected: false,
		},
		{ // #11
			rule:          config.Rule{"and": []map[string]interface{}{{"from": "@example.com"}}},
			matchExpected: true,
		},
		{ // #12
			rule:          config.Rule{"and": []map[string]interface{}{{"empty-header": ""}}},
			matchExpected: true,
		},
		{ // #13
			rule:          config.Rule{"and": []map[string]interface{}{{"subject": "löv"}}},
			matchExpected: true,
		},
		{ // #14
			rule:          config.Rule{"and": []map[string]interface{}{{"from": "@EXAMPLE.COM"}}},
			matchExpected: true,
		},
		{ // #15
			rule:          config.Rule{"and": []map[string]interface{}{{"to": "@example.com"}}},
			matchExpected: true,
		},
		{ // #16
			rule: config.Rule{"and": []map[string]interface{}{
				{"subject": "löve$"},
				{"subject": "^with löve$"},
				{"subject": "^wit.*ve$"},
				{"subject": "^with\\s+löve$"},
				{"subject": "^.*$"},
				{"subject": ".*"},
				{"subject": "^with\\s+l(ö|ä)ve$"},
				{"subject": "^with\\s+l(?:ö|ä)ve$"},
				{"subject": "^WITH"},
			}},
			matchExpected: true,
		},
		{ // #17
			rule:          config.Rule{"and": []map[string]interface{}{{"to": "!^\\ü^@example.com"}}}, // bad regex
			matchExpected: false,
			err:           "error parsing regexp: invalid escape sequence: `\\ü`",
		},
	}

	// map[string][]map[string]interface{}
	//headers: MailHeaders{"from": "oO"},
	/*
			rule: config.Rule{
			"or": {
				config.Rule{
					"or":
					[]map[string]interface{}{
						{"from": "foo"},
						{"from": "fOo"},
					},
				},
			},
		},
	*/

	testMailHeaders = MailHeaders{"from": "foo@example.com", "to": "me@EXAMPLE.com", "subject": "With Löve", "empty-header": ""}

	for i, test := range ruleParserTests {
		matched, err := ParseRule(test.rule, testMailHeaders)
		if test.err != "" && err != nil {
			require.True(strings.HasPrefix(err.Error(), test.err), "Actual error message: %v", err.Error())
		}
		require.Equal(test.matchExpected, matched, "Test #%v from ruleParserTests failed: testRule=%q testMailHeaders=%q", i+1, test.rule, testMailHeaders)
	}
}
