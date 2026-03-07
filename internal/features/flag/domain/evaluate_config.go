package domain

import (
	"fmt"
	"hash/fnv"
)

type Config struct {
	Default bool
	Rules   []Rule
}

type Rule struct {
	Conditions []Condition
	Action     Action
}

type Condition struct {
	Attribute Attribute
	Operator  Operator
	Value     string
}

type Attribute string
type Operator string

type Action struct {
	Rollout *int
	Force   *bool
}

type ActionType string
const (
	ActionTypeRollout ActionType = "rollout"
)

type Field string
const (
	RolloutKeyField Field = "rollout_key"
	FlagKeyField    Field = "flag_key"
)

type MissingFieldError struct {
	Fields []Field
}

func (e MissingFieldError) Error() string {
	return fmt.Sprintf("missing fields: %v", e.Fields)
}

func (r *Rule) Evaluate(eval EvaluationContext) (result, match bool, err error) {
	mfe := MissingFieldError{
		Fields: make([]Field, 0),
	}
	
	if eval.RolloutKey == "" {
		mfe.Fields = append(mfe.Fields, RolloutKeyField)
	}
	if eval.FlagKey == "" {
		mfe.Fields = append(mfe.Fields, FlagKeyField)
	}

	if len(mfe.Fields) > 0 {
		return false, false, mfe
	}

	if r.Action.Force != nil {
		return *r.Action.Force, true, nil
	}

	if r.Action.Rollout != nil {
		b := bucket(eval.FlagKey, eval.RolloutKey)

		return rolloutAccept(b, *r.Action.Rollout), true, nil
	}

	return false, false, fmt.Errorf("missing action")
}

func bucket(flagKey, rolloutKey string) int {
	h := fnv.New32a()

	h.Write([]byte(flagKey))
	h.Write([]byte(":"))
	h.Write([]byte(rolloutKey))

	return int(h.Sum32() % 100)
}

func rolloutAccept(bucket, rollout int) bool {
	return bucket < rollout
}