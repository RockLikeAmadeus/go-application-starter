// Package plans is the config adapter for ports.Entitlements: it reads the
// plans file, which defines each plan's limits.
package plans

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// Catalog holds each plan's entitlements, by plan name.
type Catalog map[string]domain.Entitlements

// document is the plans file's shape. Its records are separate from the
// domain types, so the file's format can change on its own.
type document struct {
	Plans map[string]planRecord `json:"plans"`
}

type planRecord struct {
	Limits map[string]limitRecord `json:"limits"`
}

type limitRecord struct {
	Max      int              `json:"max"`
	WhenOver domain.OverLimit `json:"whenOver"`
}

// Parse reads a plans document:
//
//	{"plans": {"<plan>": {"limits": {"<limit>": {"max": 3, "whenOver": "block-new"}}}}}
//
// and returns each plan's entitlements. Lines whose first non-blank characters
// are // are comments, so the file can explain itself and carry the template's
// sample markers. Comments at the end of a line aren't supported.
func Parse(data []byte) (Catalog, error) {
	var doc document
	dec := json.NewDecoder(bytes.NewReader(stripComments(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		return nil, fmt.Errorf("plans file: %w", err)
	}
	catalog := Catalog{}
	for name, plan := range doc.Plans {
		entitlements, err := plan.entitlements(name)
		if err != nil {
			return nil, fmt.Errorf("plans file: plan %q, %w", name, err)
		}
		catalog[name] = entitlements
	}
	return catalog, nil
}

// entitlements builds the named plan's entitlements, validating each limit.
func (p planRecord) entitlements(name string) (domain.Entitlements, error) {
	limits := map[string]domain.Limit{}
	for limit, l := range p.Limits {
		built, err := domain.NewLimit(limit, l.Max, l.WhenOver)
		if err != nil {
			return domain.Entitlements{}, fmt.Errorf("limit %q: %w", limit, err)
		}
		limits[limit] = built
	}
	return domain.Entitlements{Plan: name, Limits: limits}, nil
}

// stripComments empties every full-line // comment. It keeps the lines, so a
// JSON error's position still matches the file.
func stripComments(data []byte) []byte {
	lines := bytes.Split(data, []byte("\n"))
	for i, line := range lines {
		if bytes.HasPrefix(bytes.TrimSpace(line), []byte("//")) {
			lines[i] = nil
		}
	}
	return bytes.Join(lines, []byte("\n"))
}

// Entitlements is the ports.Entitlements adapter over a Catalog.
type Entitlements struct {
	// Catalog holds the plans, usually from Parse.
	Catalog Catalog
	// PlanOf names the plan a workspace is on. The local edition answers its
	// one plan for every workspace; the acceptance tests answer from a map.
	// Changing a workspace's plan (ChangePlan) will store the answers instead.
	PlanOf func(ctx context.Context, workspace domain.WorkspaceID) (string, error)
}

var _ ports.Entitlements = Entitlements{}

// For returns the entitlements of the plan the workspace is on. It returns
// PlanOf's error as is, and refuses a plan the catalog doesn't have.
func (e Entitlements) For(ctx context.Context, workspace domain.WorkspaceID) (domain.Entitlements, error) {
	plan, err := e.PlanOf(ctx, workspace)
	if err != nil {
		return domain.Entitlements{}, err
	}
	entitlements, ok := e.Catalog[plan]
	if !ok {
		return domain.Entitlements{}, fmt.Errorf("plan %q isn't in the plans file", plan)
	}
	return entitlements, nil
}
