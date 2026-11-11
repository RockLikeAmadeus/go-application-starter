package domain

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("The requested record could not be found.")

// ErrNoWorkspace is returned when a request has no workspace to work in.
var ErrNoWorkspace = errors.New("Please sign in to continue.")

// ErrValidation is the sentinel matched by every *ValidationError via errors.Is.
var ErrValidation = errors.New("validation failed")

// ValidationError describes why an input violates a business rule.
type ValidationError struct {
	// Field is the input it refuses ("name"). The API and front ends use it to
	// tell which input was at fault.
	Field string
	// Message is what people read, as a whole sentence.
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Is lets callers use errors.Is(err, ErrValidation).
func (e *ValidationError) Is(target error) bool {
	return target == ErrValidation
}

// LimitError is returned when a plan's limit refuses a change.
type LimitError struct {
	// Limit names the limit. Messages use it as the plural noun for what is
	// counted ("items"), so name limits that way.
	Limit string
	// Max is the plan's limit.
	Max int
	// Plan names the plan that sets the limit.
	Plan string
	// IsReadOnly is set when a change to existing data was refused because the
	// workspace is over a ReadOnly limit. Otherwise, creating was refused at
	// the limit.
	IsReadOnly bool
}

func (e *LimitError) Error() string {
	if e.IsReadOnly {
		return fmt.Sprintf("Your %s are read-only because you have exceeded the limit of %d %s on the %s plan.", e.Limit, e.Max, e.Limit, e.Plan)
	}
	return fmt.Sprintf("You have reached the limit of %d %s on the %s plan.", e.Max, e.Limit, e.Plan)
}

// FeatureError is returned when the workspace's plan refuses a feature.
type FeatureError struct {
	// Feature is the feature's key ("item-export").
	Feature string
	// Name is what people read ("Item export"), or the key when the plans
	// don't know the feature.
	Name string
	// AvailableOn names the plans that have the feature, for the upgrade
	// hint. It's empty when there's no hint to give.
	AvailableOn []string
}

func (e *FeatureError) Error() string {
	refusal := e.Name + " is not included in your current plan."
	if len(e.AvailableOn) == 0 {
		return refusal
	}
	upgrade := refusal + " To use this feature, please upgrade to "
	if len(e.AvailableOn) == 1 {
		return upgrade + PlansPhrase(e.AvailableOn) + "."
	}
	return upgrade + "one of the following plans: " + PlanList(e.AvailableOn) + "."
}

// PlansPhrase names plans for an "available on" hint: "the pro plan", or "the
// following plans: pro or team". plans must not be empty.
func PlansPhrase(plans []string) string {
	if len(plans) == 1 {
		return "the " + plans[0] + " plan"
	}
	return "the following plans: " + PlanList(plans)
}

// PlanList joins two or more plan names: "pro or team", or "pro, team, or
// business".
func PlanList(plans []string) string {
	last := len(plans) - 1
	if last == 1 {
		return plans[0] + " or " + plans[1]
	}
	return strings.Join(plans[:last], ", ") + ", or " + plans[last]
}
