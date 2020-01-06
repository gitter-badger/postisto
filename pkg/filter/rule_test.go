package filter

import (
	"github.com/arnisoph/postisto/pkg/config"
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
			err:           `Rule operator "non-existent-op" is unsupported`,
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
								{"subject": "^with\\s+l(?:ö|ä)ve$"},
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
						{"or": []map[string]interface{}{{"subject": "löasdv"}}},
						{"and": []map[string]interface{}{{"from": ""}}},
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

	testMailHeaders := MailHeaders{"from": "foo@example.com", "to": "me@EXAMPLE.com", "subject": "With Löve", "empty-header": ""}

	cfg := config.New()
	cfg, err := cfg.Load("../../test/data/configs/valid/test/TestParserRuleSet.yaml")
	require.Nil(err)

	acc := cfg.Accounts["test"]
	require.NotNil(acc)

	for i, test := range ruleParserTests {
		//yml, _ := yaml.Marshal(test.filters)
		//fmt.Println(string(yml))

		for filterName, filter := range test.filters {

			// Test with native synthetic test data
			matched, err := ParseRuleSet(filter.RuleSet, testMailHeaders)
			if test.err != "" && err != nil {
				require.True(strings.HasPrefix(err.Error(), test.err), "NATIVE DATA TEST: Actual error message: %v", err.Error())
			}
			require.Equal(test.matchExpected, matched, "NATIVE DATA TEST: Test #%v (%q) from ruleParserTests failed: ruleSet=%q testMailHeaders=%q", i+1, filterName, filter.RuleSet, testMailHeaders)

			// Test with same synthetic ruleSet test data from YAML
			require.Contains(acc.Filters.Names(), filterName)
			ymlFilter := acc.Filters[filterName]
			require.NotNil(ymlFilter)
			require.NotNil(ymlFilter.RuleSet)

			matched, err = ParseRuleSet(ymlFilter.RuleSet, testMailHeaders)
			if test.err != "" && err != nil {
				require.True(strings.HasPrefix(err.Error(), test.err), "YML DATA TEST: Actual error message: %v", err.Error())
			}
			require.Equal(test.matchExpected, matched, "YML DATA TEST: Test #%v (%q) from ruleParserTests failed: ruleSet=%q testMailHeaders=%q", i+1, filterName, ymlFilter.RuleSet, testMailHeaders)
		}
	}
}
