// Package jsonfile is a plain-text storage adapter that keeps one JSON file
// per domain type ("one file ≈ one table") inside a data directory:
//
//	<dir>/items.json   {"version": 1, "records": {"<id>": {...}, ...}}
//	<dir>/.lock        advisory lock held while a Store is open
//
// Trade-offs (see docs/adr/0003-plain-text-storage.md):
//   - every write rewrites the whole file, so it suits small datasets;
//   - only one process may open a data directory at a time (enforced by the
//     lock file), so do not run more than one server instance per volume.
package jsonfile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

// ErrLocked is returned by Open when another process holds the data directory.
var ErrLocked = errors.New("data directory is locked by another process")

const fileFormatVersion = 1

// Store owns a data directory and its lock. Repositories for individual
// domain types are created from it and share its mutex.
type Store struct {
	dir    string
	mu     sync.RWMutex //nolint:unused // shared by the repositories built on this Store
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

type tableFile[R any] struct {
	Version int          `json:"version"`
	Records map[string]R `json:"records"`
}

// readTable loads <dir>/<name>.json; a missing file is an empty table.
func readTable[R any](s *Store, name string) (map[string]R, error) {
	data, err := os.ReadFile(s.path(name))
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]R{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", name, err)
	}
	var tf tableFile[R]
	if err := json.Unmarshal(data, &tf); err != nil {
		return nil, fmt.Errorf("decode %s: %w", name, err)
	}
	if tf.Version != fileFormatVersion {
		return nil, fmt.Errorf("%s: unsupported file version %d", name, tf.Version)
	}
	if tf.Records == nil {
		tf.Records = map[string]R{}
	}
	return tf.Records, nil
}

// writeTable atomically replaces <dir>/<name>.json: write a temp file in the
// same directory, fsync it, then rename over the original.
func writeTable[R any](s *Store, name string, records map[string]R) error {
	data, err := json.MarshalIndent(tableFile[R]{Version: fileFormatVersion, Records: records}, "", "  ")
	if err != nil {
		return fmt.Errorf("encode %s: %w", name, err)
	}
	tmp, err := os.CreateTemp(s.dir, name+".*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename

	if _, err := tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", name, err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("sync %s: %w", name, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close %s: %w", name, err)
	}
	if err := os.Rename(tmp.Name(), s.path(name)); err != nil {
		return fmt.Errorf("replace %s: %w", name, err)
	}
	return nil
}

func (s *Store) path(name string) string {
	return filepath.Join(s.dir, name+".json")
}
