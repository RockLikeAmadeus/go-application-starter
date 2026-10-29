package domain

import (
	"errors"
	"fmt"
)

// ErrNotFound is returned when a requested entity does not exist.
var ErrNotFound = errors.New("not found")

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
