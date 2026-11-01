// Package plans is the config adapter for ports.Entitlements: it reads the
// plans file, which defines each plan's limits.
package plans

import (
	"bytes"
	"encoding/json"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
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
	_ = json.Unmarshal(stripComments(data), &doc)
	catalog := Catalog{}
	for name, plan := range doc.Plans {
		limits := map[string]domain.Limit{}
		for limit, l := range plan.Limits {
			limits[limit] = domain.Limit{Name: limit, Max: l.Max, WhenOver: l.WhenOver}
		}
		catalog[name] = domain.Entitlements{Plan: name, Limits: limits}
	}
	return catalog, nil
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
