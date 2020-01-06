package filter

import (
	"fmt"
	"github.com/arnisoph/postisto/pkg/config"
	"strings"
)

type MailHeaders map[string]string

type UnknownCommandTypeError struct {
	opName string
}

func (err *UnknownCommandTypeError) Error() string {
	return fmt.Sprintf("Rule operator %q is unsupported", err.opName)
}

func ParseRuleSet(ruleSet config.RuleSet) error {
	var err error

	fmt.Println(ruleSet)

	return err
}

func ParseRule(rule config.Rule, headers MailHeaders) (bool, error) {
	var err error
	/*
		switch r := rule.(type) {
		case config.Rule:*/
	for op, patterns := range rule {
		op = strings.ToLower(op)

		switch op {
		case "or":
			for _, pattern := range patterns {
				for patternHeaderName, patternValue := range pattern {
					//if patternHeaderName == "and" || patternHeaderName == "or" {
					//	if matched, err := ParseRule(patternValue, headers); err != nil {
					//		return false, err
					//	} else if matched {
					//		return true, err
					//	}
					//}
					//fmt.Println(patternValue, "=d=", headers[patternHeaderName])

					//switch p := patternValue.(type) {
					//case []map[string]interface{}:
					//	fmt.Println("straight")
					//	if matched, err := ParseRule(p, headers); err != nil {
					//		return false, err
					//	} else if matched {
					//		return true, err
					//	}
					//case string:
					//fmt.Println(patternValue, "=s=", headers[patternHeaderName])
					if patternValue == headers[patternHeaderName] {
						return true, nil
					}
				}
			}
		case "and":
			for _, pattern := range patterns {
				fmt.Println(pattern, "pat")
				for patternHeaderName, patternValue := range pattern {
					fmt.Println(patternValue, "patVal")
					if patternValue != headers[patternHeaderName] {
						return false, nil
					}
				}
			}

			return true, nil
		default:
			return false, &UnknownCommandTypeError{opName: op}
		}
	}
	/*case []map[string]interface{}:
	fmt.Println("hoora")
	for left, right := range r {
		fmt.Println(left, right)
	}*/
	/*
		default:
			fmt.Println("fuck!", r)
		}
	*/
	return false, err
}
