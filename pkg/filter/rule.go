package filter

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"reflect"
	"regexp"
	"strings"
)

type MailHeaders map[string]string

func ParseRuleSet(ruleSet config.RuleSet, headers MailHeaders) (bool, error) {
	var err error

	for _, rule := range ruleSet {
		//fmt.Println("rule:", rule)
		matched, err := parseRule(rule, headers)
		if err != nil {
			return false, err
		}

		if matched {
			return true, err
		}
	}

	return false, err
}

func parseRule(rule config.Rule, headers MailHeaders) (bool, error) {
	var err error

	for op, patterns := range rule {
		op = strings.ToLower(op)

		//fmt.Println("new", rule)
		switch op {
		case "or":
			for _, pattern := range patterns {
				for patternHeaderName, patternValues := range pattern {
					if matched, err := checkRulePattern(patternValues, headers[patternHeaderName]); err != nil {
						return false, err
					} else if matched {
						return true, nil
					}
				}
			}
		case "and":
			for _, pattern := range patterns {
				for patternHeaderName, patternValues := range pattern {
					if matched, err := checkRulePattern(patternValues, headers[patternHeaderName]); err != nil {
						return false, err
					} else if !matched {
						return false, nil
					}
				}
			}

			return true, nil
		default:
			return false, fmt.Errorf("rule operator %q is unsupported", op)
		}
	}

	return false, err
}

func checkRulePattern(patternValues interface{}, header string) (bool, error) {
	parsedValues, err := parsePatternValues(patternValues)
	if err != nil {
		return false, err
	}

	//fmt.Println("Values:", parsedValues)
	for _, patternValue := range parsedValues {
		matched, err := checkMatch(patternValue, header)
		if err != nil {
			return false, err
		}

		if matched {
			return true, nil
		}
	}

	return false, nil
}

func checkMatch(pattern string, s string) (bool, error) {
	pattern = strings.ToLower(pattern)
	s = strings.ToLower(s)
	var err error

	//fmt.Printf("Comparing %v with %v\n", pattern, s)

	if pattern == "" && s == "" {
		return true, err
	}

	if pattern == "" && s != "" {
		return false, err
	}

	if pattern == s {
		return true, err
	}

	if strings.Contains(s, pattern) {
		return true, err
	}

	regEx, err := regexp.Compile(pattern)
	if err != nil {
		return false, err
	}

	if regEx.MatchString(s) {
		return true, err
	}

	return false, err
}

func parsePatternValues(patternValues interface{}) ([]string, error) {
	var values []string

	switch v := patternValues.(type) {
	case string:
		return append(values, v), nil
	case int:
		return append(values, fmt.Sprintf("%v", v)), nil
	//case float32:
	//	return append(values, fmt.Sprintf("%v", v)), nil
	//case float64:
	//	return append(values, fmt.Sprintf("%v", v)), nil
	//case bool:
	//	return append(values, fmt.Sprintf("%v", v)), nil
	case []string:
		for _, val := range v {
			values = append(values, val)
		}
	case []interface{}:
		for _, val := range v {
			p, err := parsePatternValues(val)

			if err != nil {
				return values, err
			}

			values = append(values, p...)
		}
	default:
		return values, fmt.Errorf("unsupported value type %v", reflect.TypeOf(v))
	}

	return values, nil
}
