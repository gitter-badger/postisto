package filter

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"regexp"
	"strings"
)

type MailHeaders map[string]string

type UnknownCommandTypeError struct {
	opName string
}

func (err *UnknownCommandTypeError) Error() string {
	return fmt.Sprintf("Rule operator %q is unsupported", err.opName)
}

func ParseRuleSet(ruleSet config.RuleSet, headers MailHeaders) (bool, error) {
	var err error

	for _, rule := range ruleSet {
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

		switch op {
		case "or":
			for _, pattern := range patterns {
				for patternHeaderName, patternValue := range pattern {
					matched, err := checkMatch(patternValue.(string), headers[patternHeaderName])
					if err != nil {
						return false, err
					}

					if matched {
						return true, nil
					}
				}
			}
		case "and":
			for _, pattern := range patterns {
				for patternHeaderName, patternValue := range pattern {
					matched, err := checkMatch(patternValue.(string), headers[patternHeaderName])
					if err != nil {
						return false, err
					}

					if !matched {
						return false, nil
					}
				}
			}

			return true, nil
		default:
			return false, &UnknownCommandTypeError{opName: op}
		}
	}

	return false, err
}

func checkMatch(pattern string, s string) (bool, error) {
	pattern = strings.ToLower(pattern)
	s = strings.ToLower(s)
	var err error

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
