package project

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// RecordFile holds the names a project goes by, relative to its root. Runs
// write it with WriteRecord rather than renaming its text.
const RecordFile = "project.json"

// record is how Names are stored. The environment variable prefix is left out:
// it is derived from the slug.
type record struct {
	Module string `json:"module"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	App    string `json:"app"`
}

// ReadRecord reads the names the project in dir goes by from its project.json.
func ReadRecord(dir string) (Names, error) {
	data, err := os.ReadFile(filepath.Join(dir, RecordFile))
	if err != nil {
		return Names{}, err
	}
	var r record
	if err := json.Unmarshal(data, &r); err != nil {
		return Names{}, err
	}
	return Names{Module: r.Module, Brand: r.Name, Slug: r.Slug, App: r.App, EnvPrefix: EnvPrefix(r.Slug)}, nil
}

// WriteRecord records the names the project in dir goes by in its project.json.
func WriteRecord(dir string, n Names) error {
	data, err := json.MarshalIndent(record{Module: n.Module, Name: n.Brand, Slug: n.Slug, App: n.App}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, RecordFile), append(data, '\n'), 0o644)
}
