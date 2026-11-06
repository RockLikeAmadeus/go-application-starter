package jsonfile

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// Run these with task bench. The files are in the page cache, so the times are
// for reading data this computer already holds, not for a first read after a
// restart.

var benchmarkSizes = []int{1_000, 10_000}

// storeWithRecords opens a store holding n records in one table. It writes the
// files directly: writeRecord's fsync for each one would take minutes.
func storeWithRecords(b *testing.B, n int) *Store {
	b.Helper()
	s, err := Open(b.TempDir())
	if err != nil {
		b.Fatalf("Open: %v", err)
	}
	b.Cleanup(func() { s.Close() })
	dir := filepath.Join(s.dir, "things")
	if err := os.Mkdir(dir, 0o755); err != nil {
		b.Fatalf("creating the table's folder: %v", err)
	}
	for i := range n {
		key := strconv.Itoa(i)
		file := recordFile[record]{Version: recordFormatVersion, Key: key, Record: record{Name: "record " + key}}
		data, err := json.Marshal(file)
		if err != nil {
			b.Fatalf("encoding record %s: %v", key, err)
		}
		if err := os.WriteFile(filepath.Join(dir, recordFileName(key)), data, 0o644); err != nil {
			b.Fatalf("writing record %s: %v", key, err)
		}
	}
	return s
}

func BenchmarkReadRecords(b *testing.B) {
	for _, n := range benchmarkSizes {
		b.Run(fmt.Sprintf("records=%d", n), func(b *testing.B) {
			s := storeWithRecords(b, n)
			for b.Loop() {
				got, err := readRecords[record](s, "things")
				if err != nil {
					b.Fatalf("readRecords error = %v", err)
				}
				if len(got) != n {
					b.Fatalf("readRecords read %d records, want %d", len(got), n)
				}
			}
		})
	}
}

func BenchmarkReadRecord(b *testing.B) {
	for _, n := range benchmarkSizes {
		b.Run(fmt.Sprintf("records=%d", n), func(b *testing.B) {
			s := storeWithRecords(b, n)
			key := strconv.Itoa(n / 2)
			for b.Loop() {
				if _, err := readRecord[record](s, "things", key); err != nil {
					b.Fatalf("readRecord error = %v", err)
				}
			}
		})
	}
}

func BenchmarkConflictCopies(b *testing.B) {
	for _, n := range benchmarkSizes {
		b.Run(fmt.Sprintf("records=%d", n), func(b *testing.B) {
			s := storeWithRecords(b, n)
			for b.Loop() {
				copies, err := s.ConflictCopies()
				if err != nil {
					b.Fatalf("ConflictCopies error = %v", err)
				}
				if len(copies) != 0 {
					b.Fatalf("ConflictCopies = %v, want none", copies)
				}
			}
		})
	}
}
