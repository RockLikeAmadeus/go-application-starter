//go:build windows

package jsonfile

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

func lockDir(path string) (func() error, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open lock file: %w", err)
	}
	h := windows.Handle(f.Fd())
	ol := new(windows.Overlapped)
	if err := windows.LockFileEx(h, windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY, 0, 1, 0, ol); err != nil {
		f.Close()
		if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
			return nil, ErrLocked
		}
		return nil, fmt.Errorf("lock data dir: %w", err)
	}
	return func() error {
		_ = windows.UnlockFileEx(h, 0, 1, 0, ol)
		return f.Close()
	}, nil
}
