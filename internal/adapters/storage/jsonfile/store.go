// Package jsonfile is a plain-text storage adapter that keeps one JSON file
// per domain type ("one file ≈ one table") inside a data directory:
//
//	<dir>/<table>.json   {"version": 1, "records": {"<key>": {...}, ...}}
//	<dir>/.lock          advisory lock held while a Store is open
//
// Every record belongs to a workspace, and repositories only return records
// in the workspace they are asked about.
//
// Trade-offs (see docs/adr/0003-plain-text-storage.md):
//   - every write rewrites the whole file, so it suits small datasets;
//   - only one process may open a data directory at a time (enforced by the
//     lock file), so do not run more than one server instance per volume.
package jsonfile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// ErrLocked is returned by Open when another process holds the data directory.
var ErrLocked = errors.New("data directory is locked by another process")

const recordFormatVersion = 2

// Store owns a data directory and its lock. Repositories for individual
// domain types are created from it and share its mutex.
type Store struct {
	dir    string
	mu     sync.RWMutex // shared by the repositories built on this Store
	unlock func() error
}

// Open creates dir if needed and takes an exclusive lock on it.
func Open(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}
	unlock, err := lockDir(filepath.Join(dir, ".lock"))
	if err != nil {
		return nil, err
	}
	return &Store{dir: dir, unlock: unlock}, nil
}

// Close releases the directory lock.
func (s *Store) Close() error {
	return s.unlock()
}

// Ready reports whether the store can still write to its data directory,
// by creating and removing an empty temp file there. It fails when the
// directory has gone, is read-only or is full, and checks afresh each time,
// so a directory that comes back is ready again.
func (s *Store) Ready(context.Context) error {
	f, err := os.CreateTemp(s.dir, ".ready-*.tmp")
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Remove(f.Name())
}

// recordFile is the on-disk shape of one record. The key is in the file, as
// well as in its name, because not every record holds its own.
type recordFile[R any] struct {
	Version int    `json:"version"`
	Key     string `json:"key"`
	Record  R      `json:"record"`
}

// writeRecord atomically replaces <dir>/<table>/<key>.json.
func writeRecord[R any](s *Store, table, key string, rec R) error {
	data, err := json.MarshalIndent(recordFile[R]{Version: recordFormatVersion, Key: key, Record: rec}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s/%s: %w", table, key, err)
	}
	dir := filepath.Join(s.dir, table)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create %s: %w", table, err)
	}
	return replaceFile(filepath.Join(dir, recordFileName(key)), data)
}

// recordFileName is the name of the file holding the record with this key.
// Escaping keeps it a single file name on every platform: url.PathEscape
// wouldn't do, as it leaves the ":" that Windows refuses.
func recordFileName(key string) string {
	return url.QueryEscape(key) + ".json"
}

// readRecords loads every record in <dir>/<table>, by key; a missing folder
// is an empty table. It skips anything that isn't a .json file, and conflict
// copies: a copy keeps the original's contents, so its key doesn't match its
// name, whichever way a sync service named it.
func readRecords[R any](s *Store, table string) (map[string]R, error) {
	dir := filepath.Join(s.dir, table)
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]R{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", table, err)
	}
	records := make(map[string]R, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		file, err := readRecordFile[R](s, table, entry.Name())
		if err != nil {
			return nil, err
		}
		if file.Version != recordFormatVersion {
			return nil, fmt.Errorf("%s/%s: unsupported file version %d", table, entry.Name(), file.Version)
		}
		if entry.Name() != recordFileName(file.Key) {
			continue
		}
		records[file.Key] = file.Record
	}
	return records, nil
}

// readRecord loads the one record with this key from its own file, or
// returns domain.ErrNotFound if there is no such file.
func readRecord[R any](s *Store, table, key string) (R, error) {
	file, err := readRecordFile[R](s, table, recordFileName(key))
	if errors.Is(err, fs.ErrNotExist) {
		return file.Record, domain.ErrNotFound
	}
	return file.Record, err
}

// readRecordFile reads and decodes one file in <dir>/<table>.
func readRecordFile[R any](s *Store, table, name string) (recordFile[R], error) {
	var file recordFile[R]
	data, err := os.ReadFile(filepath.Join(s.dir, table, name))
	if err != nil {
		return file, fmt.Errorf("read %s/%s: %w", table, name, err)
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return file, fmt.Errorf("decode %s/%s: %w", table, name, err)
	}
	return file, nil
}

// deleteRecord removes <dir>/<table>/<key>.json.
func deleteRecord(s *Store, table, key string) error {
	if err := os.Remove(filepath.Join(s.dir, table, recordFileName(key))); err != nil {
		return fmt.Errorf("delete %s/%s: %w", table, key, err)
	}
	return nil
}

// replaceFile atomically writes data, and a final newline, to path: write a
// temp file in the same directory, fsync it, then rename over the original.
func replaceFile(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename

	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", path, err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync %s: %w", path, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", path, err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("replace %s: %w", path, err)
	}
	return nil
}
