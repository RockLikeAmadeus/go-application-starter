// Package plans is the config adapter for ports.EntitlementsProvider: it reads the
// plans file, which defines each plan's limits.
package plans

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// Catalog holds each plan's entitlements, by plan name.
type Catalog map[string]domain.Entitlements

// File is what a plans file holds: its plans, and the one a new workspace
// starts on.
type File struct {
	Catalog          Catalog
	NewWorkspacePlan string
}

// document is the plans file's shape. Its records are separate from the
// domain types, so the file's format can change on its own.
type document struct {
	NewWorkspacePlan string                   `json:"newWorkspacePlan"`
	Features         map[string]featureRecord `json:"features"`
	Plans            map[string]planRecord    `json:"plans"`
}

type featureRecord struct {
	Name    string                    `json:"name"`
	WhenOff domain.FeatureOffBehavior `json:"whenOff"`
}

type planRecord struct {
	Limits   map[string]limitRecord `json:"limits"`
	Features []string               `json:"features"`
}

type limitRecord struct {
	Max      int                      `json:"max"`
	WhenOver domain.OverLimitBehavior `json:"whenOver"`
}

// Parse reads a plans document:
//
//	{
//	  "newWorkspacePlan": "<plan>",
//	  "features": {"<feature>": {"name": "Item export", "whenOff": "show-locked"}},
//	  "plans": {"<plan>": {
//	    "limits": {"<limit>": {"max": 3, "whenOver": "block-new"}},
//	    "features": ["<feature>"]
//	  }}
//	}
//
// and returns each plan's entitlements, in File's Catalog, and the plan a new
// workspace starts on, in NewWorkspacePlan. A file without newWorkspacePlan, or
// whose newWorkspacePlan isn't one of its plans, is refused. Every plan gets
// every feature in "features", on when the plan lists it and off otherwise; a
// plan listing a feature that isn't there is refused.
// Lines whose first non-blank characters are // are comments, so the file can
// explain itself and carry the template's sample markers. Comments at the end
// of a line aren't supported.
func Parse(data []byte) (File, error) {
	var doc document
	dec := json.NewDecoder(bytes.NewReader(stripComments(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		return File{}, fmt.Errorf("plans file: %w", err)
	}
	features, err := buildFeatures(doc.Features)
	if err != nil {
		return File{}, fmt.Errorf("plans file: %w", err)
	}
	catalog := Catalog{}
	for name, plan := range doc.Plans {
		entitlements, err := plan.entitlements(name, features)
		if err != nil {
			return File{}, fmt.Errorf("plans file: plan %q, %w", name, err)
		}
		catalog[name] = entitlements
	}
	if doc.NewWorkspacePlan == "" {
		return File{}, errors.New("plans file: no newWorkspacePlan")
	}
	if _, ok := catalog[doc.NewWorkspacePlan]; !ok {
		return File{}, fmt.Errorf("plans file: newWorkspacePlan %q isn't one of its plans", doc.NewWorkspacePlan)
	}
	return File{Catalog: catalog, NewWorkspacePlan: doc.NewWorkspacePlan}, nil
}

// buildFeatures validates the file's features, all of them off.
func buildFeatures(records map[string]featureRecord) (map[string]domain.Feature, error) {
	features := map[string]domain.Feature{}
	for key, record := range records {
		built, err := domain.NewFeature(key, record.Name, record.WhenOff)
		if err != nil {
			return nil, fmt.Errorf("feature %q: %w", key, err)
		}
		features[key] = built
	}
	return features, nil
}

// entitlements builds the named plan's entitlements, validating each limit,
// and turns on the features it lists among all the file's features.
func (p planRecord) entitlements(planName string, allFeatures map[string]domain.Feature) (domain.Entitlements, error) {
	limits := map[string]domain.Limit{}
	for limitName, record := range p.Limits {
		built, err := domain.NewLimit(limitName, record.Max, record.WhenOver)
		if err != nil {
			return domain.Entitlements{}, fmt.Errorf("limit %q: %w", limitName, err)
		}
		limits[limitName] = built
	}
	features := maps.Clone(allFeatures)
	for _, key := range p.Features {
		feature, ok := features[key]
		if !ok {
			return domain.Entitlements{}, fmt.Errorf("feature %q isn't in the file's features", key)
		}
		feature.IsOn = true
		features[key] = feature
	}
	return domain.Entitlements{Plan: planName, Limits: limits, Features: features}, nil
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

// Provider is the ports.EntitlementsProvider and ports.PlanCatalog adapter over
// a Catalog.
type Provider struct {
	// Catalog holds the plans, usually from Parse.
	Catalog Catalog
	// PlanOf names the plan a workspace is on. The local edition answers its
	// one plan for every workspace. StoredPlanOf answers from the plans that
	// ChangePlan stores.
	PlanOf func(ctx context.Context, workspace domain.WorkspaceID) (string, error)
	// ShouldGiveHints fills each feature's upgrade hint (AvailableOn) with the plans
	// that have it. The local edition leaves it off, so it gives no hints;
	// the acceptance tests turn it on.
	ShouldGiveHints bool
	// LocalPlan is the local edition's plan, which a hint never names.
	LocalPlan string
}

var (
	_ ports.EntitlementsProvider = Provider{}
	_ ports.PlanCatalog          = Provider{}
)

// For returns the entitlements of the plan the workspace is on, with the
// upgrade hints when ShouldGiveHints is set. It returns PlanOf's error as is, and
// refuses a plan the catalog doesn't have.
func (p Provider) For(ctx context.Context, workspace domain.WorkspaceID) (domain.Entitlements, error) {
	plan, err := p.PlanOf(ctx, workspace)
	if err != nil {
		return domain.Entitlements{}, err
	}
	entitlements, ok := p.Catalog[plan]
	if !ok {
		return domain.Entitlements{}, fmt.Errorf("plan %q isn't in the plans file", plan)
	}
	if p.ShouldGiveHints {
		entitlements.Features = p.withHints(entitlements.Features)
	}
	return entitlements, nil
}

// HasPlan answers whether the catalog has the plan.
func (p Provider) HasPlan(_ context.Context, plan string) (bool, error) {
	_, ok := p.Catalog[plan]
	return ok, nil
}

// withHints returns a copy of features, each with the upgrade hint filled
// in. It copies, so the catalog's own features stay without hints.
func (p Provider) withHints(features map[string]domain.Feature) map[string]domain.Feature {
	hinted := maps.Clone(features)
	for key, feature := range hinted {
		feature.AvailableOn = p.plansWith(key)
		hinted[key] = feature
	}
	return hinted
}

// plansWith names, sorted, the catalog's plans that turn the feature on,
// leaving out LocalPlan.
func (p Provider) plansWith(feature string) []string {
	var names []string
	for name, plan := range p.Catalog {
		if name != p.LocalPlan && plan.Features[feature].IsOn {
			names = append(names, name)
		}
	}
	slices.Sort(names)
	return names
}

// StoredPlanOf returns a Provider.PlanOf that reads each workspace's plan from
// the repository, answering fallback when it has none. Any other error is passed
// on, so a failed read never puts a workspace on the fallback plan.
func StoredPlanOf(repository ports.WorkspacePlanRepository, fallback string) func(context.Context, domain.WorkspaceID) (string, error) {
	return func(ctx context.Context, workspace domain.WorkspaceID) (string, error) {
		plan, err := repository.PlanOf(ctx, workspace)
		if errors.Is(err, domain.ErrNotFound) {
			return fallback, nil
		}
		return plan, err
	}
}

// shipped is the plans file built into the binary.
//
//go:embed plans.jsonc
var shipped []byte

// Default returns the plans file built into the binary, plans.jsonc. A
// product edits that file to set its own plans; Load reads another one
// instead (STARTER_PLANS_FILE).
func Default() (File, error) {
	return Parse(shipped)
}

// Load returns the plans file at path, or Default's when path is empty.
func Load(path string) (File, error) {
	if path == "" {
		return Default()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return File{}, err
	}
	return Parse(data)
}
