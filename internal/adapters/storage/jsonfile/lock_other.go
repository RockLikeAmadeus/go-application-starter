//go:build !unix && !windows

package jsonfile

// lockDir is a no-op on platforms that are neither Unix nor Windows.
// Single-process use is then the operator's responsibility.
func lockDir(string) (func() error, error) {
	return func() error { return nil }, nil
}
