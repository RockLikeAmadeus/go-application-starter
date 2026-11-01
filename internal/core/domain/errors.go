package domain

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

// ErrNoWorkspace is returned when a request has no workspace to work in.
var ErrNoWorkspace = errors.New("no workspace")

// ErrValidation is the sentinel matched by every *ValidationError via errors.Is.
var ErrValidation = errors.New("validation failed")

// ValidationError describes why an input violates a business rule.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s %s", e.Field, e.Message)
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
	// ReadOnly is set when a change to existing data was refused because the
	// workspace is over a ReadOnly limit. Otherwise, creating was refused at
	// the limit.
	ReadOnly bool
}

func (e *LimitError) Error() string {
	if e.ReadOnly {
		return fmt.Sprintf("%s are read only while over the limit of %d", e.Limit, e.Max)
	}
	return fmt.Sprintf("the limit of %d %s is reached", e.Max, e.Limit)
}
