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

	ruleParserTests := []struct {
		headers       MailHeaders
		rule          config.Rule
		matchExpected bool
		err           string
	}{
		{ // #1
			headers:       MailHeaders{"from": "oO"},
			rule:          config.Rule{"or": []map[string]interface{}{{"from": "oO"}}},
			matchExpected: true,
		},
		{ // #2
			headers: MailHeaders{"from": "oO"},
			rule: config.Rule{
				"or": []map[string]interface{}{
					{"from": "oO"},
					{"from": "nope"},
				}},
			matchExpected: true,
		},
		{ // #3
			headers:       MailHeaders{"from": "oO"},
			rule:          config.Rule{"or": []map[string]interface{}{{"from": "!oO"}}},
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
			headers: MailHeaders{"from": "you", "to": "me"},
			rule: config.Rule{"and": []map[string]interface{}{
				{"from": "you"},
				{"to": "me"},
			}},
			matchExpected: true,
		},
		{ // #5
			headers: MailHeaders{"from": "you", "to": "me"},
			rule: config.Rule{"and": []map[string]interface{}{
				{"from": "you"},
				{"to": "you"},
			}},
			matchExpected: false,
		},
		{ // #6
			headers: MailHeaders{"from": "you"},
			rule: config.Rule{"and": []map[string]interface{}{
				{"from": "you"},
				{"to": "you"},
			}},
			matchExpected: false,
		},
		{ // #7
			headers: MailHeaders{"from": "you"},
			rule: config.Rule{"non-existent-op": []map[string]interface{}{
				{"from": "you"},
				{"to": "you"},
			}},
			matchExpected: false,
			err:           `Rule operator "non-existent-op" is unsupported`,
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

	/*	ruleParserTests = append(ruleParserTests, struct {
			headers       MailHeaders
			rule          config.Rule
			matchExpected bool
		}{
			headers:       MailHeaders{"from": "oO"},
			rule:          rule,
			matchExpected: true,
		})*/

	for i, test := range ruleParserTests {
		matched, err := ParseRule(test.rule, test.headers)
		if test.err != "" && err != nil {
			require.True(strings.HasPrefix(err.Error(), test.err), "Actual error message: %v", err.Error())
		}
		require.Equal(test.matchExpected, matched, "Test #%v from ruleParserTests failed", i+1)
	}
}
