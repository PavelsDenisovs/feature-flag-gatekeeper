package domain

import (
	"errors"
	"fmt"
	"hash/fnv"
)

var (
	ErrIncompleteContext    = errors.New("incomplete evaluation context")
	ErrUnknownConditionKind = errors.New("unknown condition kind")
	ErrUnknownOperator      = errors.New("unknown operator")
)

const CurrentConfigVersion = 1 // Increment this when changing Config struct after MVP

type Config struct {
	Default bool
	Version int
	Rules   []Rule
}

type Rule struct {
	Conditions []Condition
	Result     bool
}

type Condition struct {
	Kind       ConditionKind
	Attribute  string
	Operator   Operator
	Value      string
	Percentage int
}

type ConditionKind string

const (
	ConditionKindAttribute ConditionKind = "attribute"
	ConditionKindRollout   ConditionKind = "rollout"
)

type Operator string

const (
	OperatorEquals Operator = "equals"
)

func (r *Rule) Evaluate(eval EvaluationContext) (result, match bool, err error) {
	for _, c := range r.Conditions {
		match, err := evaluateCondition(c, eval)
		if err != nil {
			return false, false, err
		}
		if !match {
			return false, false, nil
		}
	}

	return r.Result, true, nil
}

func bucket(flagKey, subjectKey string) int {
	h := fnv.New32a()

	h.Write([]byte(flagKey))
	h.Write([]byte(":"))
	h.Write([]byte(subjectKey))

	return int(h.Sum32() % 100)
}

func rolloutAccept(bucket, rollout int) bool {
	return bucket < rollout
}

func evaluateCondition(c Condition, eval EvaluationContext) (bool, error) {
	switch c.Kind {
	case ConditionKindRollout:
		var fields []string
		if eval.FlagKey == "" {
			fields = append(fields, "flag_key")
		}

		if eval.SubjectKey == "" {
			fields = append(fields, "subject_key")
		}

		if len(fields) > 0 {
			return false, fmt.Errorf("%w: missing fields: %v", ErrIncompleteContext, fields)
		}

		b := bucket(eval.FlagKey, eval.SubjectKey)

		return rolloutAccept(b, c.Percentage), nil
	case ConditionKindAttribute:
		v, ok := eval.Attributes[c.Attribute]
		if !ok {
			return false, fmt.Errorf("%w: attribute %s is missing", ErrIncompleteContext, c.Attribute)
		}
		switch c.Operator {
		case OperatorEquals:
			return v == c.Value, nil
		default:
			return false, fmt.Errorf("%w: %v", ErrUnknownOperator, c.Operator)
		}
	}

	return false, fmt.Errorf("%w: %v", ErrUnknownConditionKind, c.Kind)
}
