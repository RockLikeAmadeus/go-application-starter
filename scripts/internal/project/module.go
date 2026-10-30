package project

import (
	"fmt"

	"golang.org/x/mod/module"
)

// CheckModule refuses a path Go would not accept as a module path.
func CheckModule(path string) error {
	if err := module.CheckPath(path); err != nil {
		return fmt.Errorf("invalid module path: %w", err)
	}
	return nil
}
