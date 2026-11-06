package jsonfile

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
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

func TestReadingATableSkipsWhatIsNotAJSONFile(t *testing.T) {
	s := openTestStore(t)
	if err := writeRecord(s, "things", "1", record{Name: "one"}); err != nil {
		t.Fatalf("writeRecord error = %v", err)
	}
	dir := filepath.Join(s.dir, "things")
	for _, name := range []string{"1.json.123.tmp", ".1.json.icloud", ".DS_Store"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("not a record"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "folder.json"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := readRecords[record](s, "things")
	if err != nil {
		t.Fatalf("readRecords error = %v", err)
	}
	if len(got) != 1 || got["1"] != (record{Name: "one"}) {
		t.Errorf("readRecords = %v, want one record under key 1", got)
	}
}

func TestReadingATableSkipsAFileWhoseKeyDoesNotMatchItsName(t *testing.T) {
	copies := map[string]string{
		"Dropbox":      "1 (Sam's conflicted copy 2026-11-06).json",
		"Syncthing":    "1.sync-conflict-20261106-120000-ABCDEFG.json",
		"OneDrive":     "1-DESKTOP-ABC123.json",
		"Google Drive": "1 (1).json",
		"iCloud":       "1 2.json",
		"by hand":      "copy of 1.json",
	}
	for service, name := range copies {
		t.Run(service, func(t *testing.T) {
			s := openTestStore(t)
			if err := writeRecord(s, "things", "2", record{Name: "two"}); err != nil {
				t.Fatalf("writeRecord error = %v", err)
			}
			// A copy keeps the original's contents, so rename the original
			// rather than keep both: most of these names sort before
			// "1.json", which would then hide a copy that was wrongly read.
			if err := writeRecord(s, "things", "1", record{Name: "one"}); err != nil {
				t.Fatalf("writeRecord error = %v", err)
			}
			dir := filepath.Join(s.dir, "things")
			if err := os.Rename(filepath.Join(dir, "1.json"), filepath.Join(dir, name)); err != nil {
				t.Fatal(err)
			}

			got, err := readRecords[record](s, "things")
			if err != nil {
				t.Fatalf("readRecords error = %v", err)
			}
			if len(got) != 1 || got["2"] != (record{Name: "two"}) {
				t.Errorf("readRecords = %v, want only the record under key 2", got)
			}
		})
	}
}

func TestReadRecordReturnsOneRecordWithoutReadingAnyOther(t *testing.T) {
	s := openTestStore(t)
	if err := writeRecord(s, "things", "1", record{Name: "one"}); err != nil {
		t.Fatalf("writeRecord error = %v", err)
	}
	// Reading the whole table would fail on this file.
	if err := os.WriteFile(filepath.Join(s.dir, "things", "2.json"), []byte("not a record"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := readRecord[record](s, "things", "1")
	if err != nil {
		t.Fatalf("readRecord error = %v", err)
	}
	if got != (record{Name: "one"}) {
		t.Errorf("readRecord = %+v, want the record named one", got)
	}
}

func TestReadRecordOfAMissingFileIsNotFound(t *testing.T) {
	s := openTestStore(t)
	if err := writeRecord(s, "things", "1", record{Name: "one"}); err != nil {
		t.Fatalf("writeRecord error = %v", err)
	}

	_, err := readRecord[record](s, "things", "2")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("readRecord error = %v, want domain.ErrNotFound", err)
	}
}

func TestDeletingARecordRemovesOnlyItsFile(t *testing.T) {
	s := openTestStore(t)
	const key = "workspace/1"
	if err := writeRecord(s, "things", key, record{Name: "one"}); err != nil {
		t.Fatalf("writeRecord error = %v", err)
	}
	if err := writeRecord(s, "things", "2", record{Name: "two"}); err != nil {
		t.Fatalf("writeRecord error = %v", err)
	}

	if err := deleteRecord(s, "things", key); err != nil {
		t.Fatalf("deleteRecord error = %v", err)
	}

	got, err := readRecords[record](s, "things")
	if err != nil {
		t.Fatalf("readRecords error = %v", err)
	}
	if len(got) != 1 || got["2"] != (record{Name: "two"}) {
		t.Errorf("readRecords = %v, want only the record under key 2", got)
	}
}
