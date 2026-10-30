package main

import (
	"os"
	"path/filepath"
	"strings"
)

// Tags for the marker lines that fence code inside shared files (see
// RemoveBlocks): template blocks are always removed, sample blocks only with
// --remove-sample.
const (
	templateTag = "template"
	sampleTag   = "sample"
)

// toolFiles are the init tool and its documentation, which every run deletes:
// a project only needs making its own once.
var toolFiles = []string{
	"scripts/initproject/",
	"docs/template.md",
}

// SampleFiles lists the template's files and directories (trailing slash) that
// exist only for the sample Item domain, relative to the repository root.
// --remove-sample deletes them outright. Item code inside shared files is
// fenced with sample:start / sample:end markers instead (see RemoveBlocks).
var SampleFiles = []string{
	// Core
	"internal/core/domain/item.go",
	"internal/core/domain/item_test.go",
	"internal/core/app/item_service.go",
	"internal/core/app/item_service_test.go",

	// Storage
	"internal/adapters/storage/memory/item_repository.go",
	"internal/adapters/storage/memory/item_repository_test.go",
	"internal/adapters/storage/jsonfile/item_repository.go",
	"internal/adapters/storage/jsonfile/item_repository_test.go",
	"internal/adapters/storage/storagetest/item_repository_contract.go",
	"internal/adapters/storage/storageutil/",

	// Front ends and the HTTP client they use
	"internal/adapters/cli/item.go",
	"internal/adapters/cli/item_test.go",
	"internal/adapters/httpapi/items.go",
	"internal/adapters/httpclient/item_service.go",
	"internal/adapters/httpclient/item_service_test.go",
	"internal/adapters/tui/items.go",
	"internal/adapters/tui/items_test.go",
	"web/src/features/items/",

	// Specifications
	"features/items.feature",
	"test/acceptance/item_steps.go",
	"web/e2e/features/items.feature",
	"web/e2e/steps/items.steps.ts",
}

// IsSample reports whether rel (slash-separated) is a SampleFiles entry or lies
// inside one.
func IsSample(rel string) bool {
	return isUnder(rel, SampleFiles)
}

// isUnder reports whether rel (slash-separated) is one of entries or lies
// inside one of them.
func isUnder(rel string, entries []string) bool {
	for _, entry := range entries {
		if rel == entry || strings.HasPrefix(rel, strings.TrimSuffix(entry, "/")+"/") {
			return true
		}
	}
	return false
}

// removeAll deletes each of entries under dir, along with any directories
// that leaves empty.
func removeAll(dir string, entries []string) error {
	for _, entry := range entries {
		p := filepath.Join(dir, filepath.FromSlash(entry))
		if err := os.RemoveAll(p); err != nil {
			return err
		}
		removeEmptyParents(dir, p)
	}
	return nil
}

// removeEmptyParents removes the directories above path, from the nearest up,
// until one is not empty. It stops at dir, which path must lie inside.
func removeEmptyParents(dir, path string) {
	dir = filepath.Clean(dir)
	for parent := filepath.Dir(path); parent != dir; parent = filepath.Dir(parent) {
		if os.Remove(parent) != nil { // fails once a directory is not empty
			return
		}
	}
}
