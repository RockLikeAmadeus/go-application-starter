package acceptance

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// itemExport is the sample feature's key in the plans file.
const itemExport = "item-export"

// registerPlanFeatureSteps binds the item export feature's steps in
// features/plan-features.feature, which only the drivers whose plans the tests
// set run.
func (w *world) registerPlanFeatureSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the "([^"]*)" plan (has|doesn't have) item export$`, w.thePlanHasItemExport)
	sc.Step(`^item export is (shown locked|hidden) when off$`, w.itemExportIsShownWhenOff)
	sc.Step(`^item export is (on|hidden|locked, available on the "([^"]*)" plan) for my workspace$`, w.itemExportIsForMyWorkspace)
	sc.Step(`^I export the items$`, w.iExportTheItems)
	sc.Step(`^I try to export the items$`, w.iTryToExportTheItems)
	sc.Step(`^the export lists:$`, w.theExportLists)
	sc.Step(`^I am told item export isn't on my plan(?:, and is available on the "([^"]*)" plan)?$`, w.iAmToldItemExportIsntOnMyPlan)
}

func (w *world) iExportTheItems(ctx context.Context) error {
	var err error
	w.exportedCSV, err = w.items.ExportItems(ctx)
	return err
}

func (w *world) iTryToExportTheItems(ctx context.Context) error {
	w.exportedCSV, w.lastErr = w.items.ExportItems(ctx)
	return nil
}

// theExportLists checks the export's names, in order, under its header.
func (w *world) theExportLists(table *godog.Table) error {
	records, err := csv.NewReader(bytes.NewReader(w.exportedCSV)).ReadAll()
	if err != nil {
		return fmt.Errorf("read the export %q: %w", w.exportedCSV, err)
	}
	if len(records) == 0 || !slices.Equal(records[0], []string{"name", "createdAt"}) {
		return fmt.Errorf("the export has no name,createdAt header: %q", w.exportedCSV)
	}
	var got, want []string
	for _, record := range records[1:] {
		got = append(got, record[0])
	}
	for _, row := range table.Rows {
		want = append(want, row.Cells[0].Value)
	}
	if !slices.Equal(got, want) {
		return fmt.Errorf("the export lists %q, want %q", got, want)
	}
	return nil
}

func (w *world) iAmToldItemExportIsntOnMyPlan(availableOn string) error {
	want := domain.FeatureError{Feature: itemExport, Name: "Item export"}
	if availableOn != "" {
		want.AvailableOn = []string{availableOn}
	}
	var got *domain.FeatureError
	if !errors.As(w.lastErr, &got) {
		return fmt.Errorf("expected %q, got %w", want.Error(), w.lastErr)
	}
	if got.Feature != want.Feature || got.Name != want.Name || !slices.Equal(got.AvailableOn, want.AvailableOn) {
		return fmt.Errorf("feature error = %+v, want %+v", *got, want)
	}
	return nil
}

// whenOff is what the scenarios call each way of showing a feature that's off.
var whenOff = map[string]domain.FeatureOffBehavior{
	"shown locked": domain.ShowLocked,
	"hidden":       domain.Hide,
}

func (w *world) thePlanHasItemExport(plan, has string) error {
	if w.plans == nil {
		return errFixedPlans
	}
	return w.plans.setFeature(plan, itemExport, has == "has")
}

func (w *world) itemExportIsShownWhenOff(shown string) error {
	if w.plans == nil {
		return errFixedPlans
	}
	w.plans.setWhenOff(itemExport, whenOff[shown])
	return nil
}

func (w *world) itemExportIsForMyWorkspace(ctx context.Context, state, availableOn string) error {
	got, err := w.entitlements.CurrentEntitlements(ctx)
	if err != nil {
		return err
	}
	f, ok := got.Features[itemExport]
	if !ok {
		return fmt.Errorf("the entitlements have no %q feature: %+v", itemExport, got.Features)
	}
	switch state {
	case "on":
		if !f.On {
			return fmt.Errorf("item export is off, want on: %+v", f)
		}
	case "hidden":
		if f.On || f.WhenOff != domain.Hide {
			return fmt.Errorf("item export = %+v, want off and hidden", f)
		}
	default:
		if f.On || f.WhenOff != domain.ShowLocked {
			return fmt.Errorf("item export = %+v, want off and shown locked", f)
		}
		if want := []string{availableOn}; !slices.Equal(f.AvailableOn, want) {
			return fmt.Errorf("item export's AvailableOn = %q, want %q", f.AvailableOn, want)
		}
	}
	return nil
}

// The harness's plans, as these steps change them. They live here, beside
// the only steps that use them, so the template without the sample has none
// unused.

// setFeature turns the feature on or off in the plan, adding the plan when
// the catalog hasn't got it. The feature must be in the catalog already (each
// plan has every feature in the plans file), so its name and whenOff stay the
// file's.
func (p *testPlans) setFeature(plan, key string, on bool) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	f, ok := p.catalogFeature(key)
	if !ok {
		return fmt.Errorf("no plan has the feature %q", key)
	}
	f.On = on
	entitlements, ok := p.catalog[plan]
	if !ok {
		entitlements = domain.Entitlements{Plan: plan, Limits: map[string]domain.Limit{}}
	}
	p.catalog[plan] = withFeature(entitlements, f)
	return nil
}

// setWhenOff changes what every plan shows of the feature when it's off.
func (p *testPlans) setWhenOff(key string, whenOff domain.FeatureOffBehavior) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for name, entitlements := range p.catalog {
		f, ok := entitlements.Features[key]
		if !ok {
			continue
		}
		f.WhenOff = whenOff
		p.catalog[name] = withFeature(entitlements, f)
	}
}

// withFeature returns the entitlements with f in place of their feature of
// the same key. It copies the features, so other holders of the map (the
// shipped catalog's plans) stay as they were.
func withFeature(entitlements domain.Entitlements, f domain.Feature) domain.Entitlements {
	features := maps.Clone(entitlements.Features)
	if features == nil {
		features = map[string]domain.Feature{}
	}
	features[f.Key] = f
	entitlements.Features = features
	return entitlements
}

// catalogFeature returns the feature as some plan has it.
func (p *testPlans) catalogFeature(key string) (domain.Feature, bool) {
	for _, entitlements := range p.catalog {
		if f, ok := entitlements.Features[key]; ok {
			return f, true
		}
	}
	return domain.Feature{}, false
}
