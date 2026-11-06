package jsonfile

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The repositories exercise these helpers too; testing them directly covers
// what no repository contract does, such as an unknown file version.

type record struct {
	Name string `json:"name"`
}

func openTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestMissingTableIsEmpty(t *testing.T) {
	got, err := readTable[record](openTestStore(t), "things")
	if err != nil {
		t.Fatalf("readTable error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("readTable = %v, want an empty table", got)
	}
}

func TestTableSurvivesAWriteAndRead(t *testing.T) {
	s := openTestStore(t)
	want := map[string]record{"1": {Name: "one"}}

	if err := writeTable(s, "things", want); err != nil {
		t.Fatalf("writeTable error = %v", err)
	}
	got, err := readTable[record](s, "things")
	if err != nil {
		t.Fatalf("readTable error = %v", err)
	}
	if len(got) != 1 || got["1"] != want["1"] {
		t.Errorf("readTable = %v, want %v", got, want)
	}
}

func TestTableWithAnUnknownVersionIsRefused(t *testing.T) {
	s := openTestStore(t)
	if err := os.WriteFile(s.path("things"), []byte(`{"version": 99, "records": {}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := readTable[record](s, "things")
	if err == nil || !strings.Contains(err.Error(), "unsupported file version 99") {
		t.Errorf("readTable error = %v, want unsupported file version 99", err)
	}
}

func TestRecordIsKeptInItsOwnVersion2FileAndReadsBack(t *testing.T) {
	s := openTestStore(t)

	if err := writeRecord(s, "things", "1", record{Name: "one"}); err != nil {
		t.Fatalf("writeRecord error = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(s.dir, "things", "1.json"))
	if err != nil {
		t.Fatalf("reading the record's file: %v", err)
	}
	var file struct {
		Version int    `json:"version"`
		Key     string `json:"key"`
		Record  record `json:"record"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("decoding the record's file: %v", err)
	}
	if file.Version != 2 || file.Key != "1" || file.Record.Name != "one" {
		t.Errorf("file = %+v, want version 2, key 1, name one", file)
	}

	got, err := readRecords[record](s, "things")
	if err != nil {
		t.Fatalf("readRecords error = %v", err)
	}
	if len(got) != 1 || got["1"] != (record{Name: "one"}) {
		t.Errorf("readRecords = %v, want one record under key 1", got)
	}
}

func TestKeyThatCannotBeAFileNameRoundTrips(t *testing.T) {
	s := openTestStore(t)
	const key = `workspace/1:a\b*c?d"e<f>g|h`

	if err := writeRecord(s, "things", key, record{Name: "one"}); err != nil {
		t.Fatalf("writeRecord error = %v", err)
	}

	entries, err := os.ReadDir(filepath.Join(s.dir, "things"))
	if err != nil {
		t.Fatalf("reading the table's folder: %v", err)
	}
	if len(entries) != 1 || entries[0].IsDir() {
		t.Fatalf("table folder holds %v, want one file", entries)
	}
	if name := entries[0].Name(); strings.ContainsAny(name, `/:\*?"<>|`) {
		t.Errorf("file name = %q, want no character a file system can refuse", name)
	}

	got, err := readRecords[record](s, "things")
	if err != nil {
		t.Fatalf("readRecords error = %v", err)
	}
	if len(got) != 1 || got[key] != (record{Name: "one"}) {
		t.Errorf("readRecords = %v, want one record under key %q", got, key)
	}
}

func TestTableWithNoRecordsFolderIsEmpty(t *testing.T) {
	got, err := readRecords[record](openTestStore(t), "things")
	if err != nil {
		t.Fatalf("readRecords error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("readRecords = %v, want an empty table", got)
	}
}

func TestRecordWithAnUnknownVersionIsRefused(t *testing.T) {
	s := openTestStore(t)
	dir := filepath.Join(s.dir, "things")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := []byte(`{"version": 99, "key": "1", "record": {"name": "one"}}`)
	if err := os.WriteFile(filepath.Join(dir, "1.json"), content, 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := readRecords[record](s, "things")
	if err == nil || !strings.Contains(err.Error(), "unsupported file version 99") {
		t.Errorf("readRecords error = %v, want unsupported file version 99", err)
	}
}
