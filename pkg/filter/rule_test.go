package filter

import (
	"github.com/arnisoph/postisto/pkg/config"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestParseRuleSet(t *testing.T) {
	require := require.New(t)

	// ACTUAL TESTS BELOW

	ruleParserTests := []struct {
		filters       config.FilterSet
		matchExpected bool
		err           string
	}{
		{
			filters: config.FilterSet{
				"simple 1o1 comparison": config.Filter{
					RuleSet: config.RuleSet{
						{
							"or": []map[string]interface{}{
								{"from": "foo@example.com"},
							},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"simple 101 comparison in or": config.Filter{
					RuleSet: config.RuleSet{
						{
							"or": []map[string]interface{}{
								{"from": "oO"},
								{"from": "foo@example.com"},
							},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"failing simple comparison": config.Filter{
					RuleSet: config.RuleSet{
						{
							"or": []map[string]interface{}{{"from": "wrong value"}},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: config.FilterSet{
				"comparison with uppercase text": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{
								{"from": "foo@example.com"},
								{"to": "me@EXAMPLE.com"},
							},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"failing and comparison": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{
								{"from": "you"},
								{"to": "you"},
							},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: config.FilterSet{
				"failing or comparison": config.Filter{
					RuleSet: config.RuleSet{
						{
							"or": []map[string]interface{}{
								{"from": "you"},
								{"to": "you"},
							},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: config.FilterSet{
				"failing and comparison": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{
								{"from": "you"},
								{"to": "you"},
							},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: config.FilterSet{
				"failing with unsupported op": config.Filter{
					RuleSet: config.RuleSet{
						{
							"non-existent-op": []map[string]interface{}{
								{"from": "you"},
								{"to": "you"},
							},
						},
					},
				},
			},
			matchExpected: false,
			err:           `rule operator "non-existent-op" is unsupported`,
		},
		{
			filters: config.FilterSet{
				"invalid value type": config.Filter{
					RuleSet: config.RuleSet{
						{
							"or": []map[string]interface{}{
								{"from": config.ConnectionConfig{}},
							},
						},
					},
				},
			},
			matchExpected: false,
			err:           `unsupported value type config.ConnectionConfig`,
		},
		{
			filters: config.FilterSet{
				"invalid nested value type": config.Filter{
					RuleSet: config.RuleSet{
						{
							"or": []map[string]interface{}{
								{"from": []interface{}{"wrong1", "wrong2", "42", []interface{}{config.ConnectionConfig{}}}},
							},
						},
					},
				},
			},
			matchExpected: false,
			err:           `unsupported value type config.ConnectionConfig`,
		},
		{
			filters: config.FilterSet{
				"substring comparison with and": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{{"from": "@example.com"}},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"substring comparison with or": config.Filter{
					RuleSet: config.RuleSet{
						{
							"or": []map[string]interface{}{{"from": "@example.com"}},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"failing on search for empty header": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{{"from": ""}},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: config.FilterSet{
				"successfully searching for empty header": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{{"empty-header": ""}},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"testing with ütf-8": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{{"subject": "löv"}},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"uppercase in rule + substring comparison": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{{"from": "@EXAMPLE.COM"}},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"uppercase in header comparison": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{{"to": "@example.com"}},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"regex comparison": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{
								{"subject": "löve$"},
								{"subject": "^with löve$"},
								{"subject": "^wit.*ve$"},
								{"subject": "^with\\s+löve$"},
								{"subject": "^.*$"},
								{"subject": ".*"},
								{"subject": "^with\\s+l(ö|ä)ve$"},
								{"suBject": "^with\\s+l(?:ö|ä)ve$"},
								{"subject": "^WITH"},
							},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"comparison with bad regex (and)": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{{"tO": "!^\\ü^@example.com"}},
						},
					},
				},
			},
			matchExpected: false,
			err:           "error parsing regexp: invalid escape sequence: `\\ü`",
		},
		{
			filters: config.FilterSet{
				"comparison with bad regex (or)": config.Filter{
					RuleSet: config.RuleSet{
						{
							"or": []map[string]interface{}{{"to": "!^\\ü^@example.com"}},
						},
					},
				},
			},
			matchExpected: false,
			err:           "error parsing regexp: invalid escape sequence: `\\ü`",
		},
		{
			filters: config.FilterSet{
				"several rules in ruleSet success": config.Filter{
					RuleSet: config.RuleSet{
						{"and": []map[string]interface{}{{"to": "@example.com"}}},
						{"or": []map[string]interface{}{{"subject": "löv"}}},
						{"and": []map[string]interface{}{{"from": ""}}},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"several rules in ruleSet failing": config.Filter{
					RuleSet: config.RuleSet{
						{"and": []map[string]interface{}{{"to": "@examplde.com"}}},
						{"or": []map[string]interface{}{{"sUbject": "löasdv"}}},
						{"and": []map[string]interface{}{{"from": ""}}},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: config.FilterSet{
				"1o1 comparison with multiple values": config.Filter{
					RuleSet: config.RuleSet{
						{
							"or": []map[string]interface{}{
								{"froM": []string{"foo@example.com", "example.com", "foo"}},
							},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"101 comparison in or with multiple values": config.Filter{
					RuleSet: config.RuleSet{
						{
							"or": []map[string]interface{}{
								{"froM": "oO"},
								{"from": []string{"foo@example.com", "example.com", "foo"}},
							},
						},
					},
				},
			},
			matchExpected: true,
		},
		{
			filters: config.FilterSet{
				"101 comparison in OR with multiple values (failing)": config.Filter{
					RuleSet: config.RuleSet{
						{
							"or": []map[string]interface{}{
								{"From": "baz"},
								{"from": []interface{}{"wrong1", "wrong2", "42", 42}},
							},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: config.FilterSet{
				"101 comparison in AND with multiple values (failing)": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{
								{"from": "baz"},
								{"from": []string{"foo@example.com", "example.com", "foo"}},
							},
						},
					},
				},
			},
			matchExpected: false,
		},
		{
			filters: config.FilterSet{
				"weirdest bug so far": config.Filter{
					RuleSet: config.RuleSet{
						{
							"and": []map[string]interface{}{
								{"X-Custom-Mail-Id": "16"},
								{"X-Notes-Item": "CSMemoFrom"},
							},
						},
					},
				},
			},
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
	}

	testMailHeaders := config.MailHeaders{"from": "foo@example.com", "to": "me@EXAMPLE.com", "subject": "With Löve", "empty-header": "", "custom-Header": "Foobar"}

	cfg := config.NewConfig()
	cfg, err := cfg.Load("../../test/data/configs/valid/test/TestParserRuleSet.yaml")
	require.Nil(err)

	acc := cfg.Accounts["test"]
	require.NotNil(acc)

	for i, test := range ruleParserTests {
		for filterName, filter := range test.filters {
			// Test with native synthetic test data
			matched, err := ParseRuleSet(filter.RuleSet, testMailHeaders)
			if test.err == "" {
				require.Nil(err)
			}
			if test.err != "" && err != nil {
				require.True(strings.HasPrefix(err.Error(), test.err), "NATIVE DATA TEST: Actual error message: %v", err.Error())
			}

			require.Equal(test.matchExpected, matched, "NATIVE DATA TEST: Test #%v (%q) from ruleParserTests failed! ruleSet=%q testMailHeaders=%q", i+1, filterName, filter.RuleSet, testMailHeaders)

			if filterName == "invalid value type" || filterName == "invalid nested value type" {
				// can't test NON-JSON data types in YAML
				continue
			}

			// Test with same synthetic ruleSet test data from YAML
			yml, err := yaml.Marshal(test.filters)
			require.Contains(acc.FilterSet.Names(), filterName, "Add test %q to TestParserRuleSet.yml:\n=========\n%v=========\n%v", filterName, string(yml), err)

			ymlFilter := acc.FilterSet[filterName]
			require.NotNil(ymlFilter)
			require.NotNil(ymlFilter.RuleSet)

			matched, err = ParseRuleSet(ymlFilter.RuleSet, testMailHeaders)
			if test.err == "" {
				require.Nil(err)
			}
			if test.err != "" && err != nil {
				require.True(strings.HasPrefix(err.Error(), test.err), "YML DATA TEST: Actual error message: %v", err.Error())
			}
			require.Equal(test.matchExpected, matched, "YML DATA TEST: Test #%v (%q) from ruleParserTests failed: ruleSet=%q testMailHeaders=%q", i+1, filterName, ymlFilter.RuleSet, testMailHeaders)
		}
	}
}
