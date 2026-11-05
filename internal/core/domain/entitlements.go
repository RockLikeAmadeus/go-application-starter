package domain

import (
	"fmt"
	"strings"
)

// OverLimitBehavior says what happens when a workspace is over a limit.
type OverLimitBehavior string

const (
	// BlockNew keeps existing data changeable and refuses only new data.
	BlockNew OverLimitBehavior = "block-new"
	// ReadOnly also refuses changes to existing data, until the workspace is
	// back within the limit. Deleting is always allowed.
	ReadOnly OverLimitBehavior = "read-only"
)

// Limit is a number a plan sets, such as how many items a workspace may have.
type Limit struct {
	Name     string
	Max      int
	WhenOver OverLimitBehavior
}

// Entitlements are what a workspace's plan allows it.
type Entitlements struct {
	Plan     string
	Limits   map[string]Limit
	Features map[string]Feature
}

// CheckCreate says whether a workspace that has count of the limited thing may
// create one more. It refuses with a *LimitError at or over the limit,
// whatever the limit's WhenOver, and allows anything when the plan has no
// such limit.
func (e Entitlements) CheckCreate(limit string, count int) error {
	l, ok := e.Limits[limit]
	if !ok || count < l.Max {
		return nil
	}
	return &LimitError{Limit: limit, Max: l.Max}
}

// CheckChange says whether a workspace that has count of the limited thing may
// change one of them. It refuses with a *LimitError (ReadOnly) only over a
// ReadOnly limit; at the limit, over a BlockNew one, or with no such limit,
// changes are allowed.
func (e Entitlements) CheckChange(limit string, count int) error {
	l, ok := e.Limits[limit]
	if !ok || l.WhenOver != ReadOnly || count <= l.Max {
		return nil
	}
	return &LimitError{Limit: limit, Max: l.Max, ReadOnly: true}
}

// NewLimit builds a valid Limit. It refuses a negative max, and a WhenOver
// other than BlockNew or ReadOnly, with a *ValidationError whose Field is the
// plans file's key ("max", "whenOver"). Build limits through it, never as a
// literal, so a bad plans file is caught when it's read.
func NewLimit(name string, max int, whenOver OverLimitBehavior) (Limit, error) {
	if max < 0 {
		return Limit{}, &ValidationError{Field: "max", Message: "must not be negative"}
	}
	if whenOver != BlockNew && whenOver != ReadOnly {
		return Limit{}, &ValidationError{Field: "whenOver", Message: fmt.Sprintf("must be %q or %q", BlockNew, ReadOnly)}
	}
	return Limit{Name: name, Max: max, WhenOver: whenOver}, nil
}

// FeatureOff says what front ends show of a feature that's off.
type FeatureOff string

const (
	// ShowLocked shows the feature as locked, with a hint naming the plans
	// that have it.
	ShowLocked FeatureOff = "show-locked"
	// Hide leaves the feature out altogether.
	Hide FeatureOff = "hide"
)

// Feature is something a plan turns on or off, such as exporting items.
type Feature struct {
	// Key is what code and the API use ("item-export").
	Key string
	// Name is what people read ("Item export").
	Name string
	// On is set when the workspace's plan has the feature.
	On      bool
	WhenOff FeatureOff
	// AvailableOn names the plans that have the feature, for the upgrade
	// hint. It's empty when the edition gives no hints.
	AvailableOn []string
}

// NewFeature builds a valid Feature. It refuses a blank name, and a WhenOff
// other than ShowLocked or Hide, with a *ValidationError whose Field is the
// plans file's key ("name", "whenOff"). Build features through it, never as a
// literal, so a bad plans file is caught when it's read.
func NewFeature(key, name string, whenOff FeatureOff) (Feature, error) {
	if strings.TrimSpace(name) == "" {
		return Feature{}, &ValidationError{Field: "name", Message: "must not be empty"}
	}
	if whenOff != ShowLocked && whenOff != Hide {
		return Feature{}, &ValidationError{Field: "whenOff", Message: fmt.Sprintf("must be %q or %q", ShowLocked, Hide)}
	}
	return Feature{Key: key, Name: name, WhenOff: whenOff}, nil
}

// CheckFeature says whether the workspace may use the feature. It refuses
// with a *FeatureError when the feature is off, and also when the plans don't
// know it, so a missing feature never means "on". Only a ShowLocked feature's
// error carries the upgrade hint (AvailableOn); a hidden or unknown one has
// none.
func (e Entitlements) CheckFeature(key string) error {
	f, ok := e.Features[key]
	if !ok {
		return &FeatureError{Feature: key, Name: key}
	}
	if f.On {
		return nil
	}
	refused := &FeatureError{Feature: key, Name: f.Name}
	if f.WhenOff == ShowLocked {
		refused.AvailableOn = f.AvailableOn
	}
	return refused
}
