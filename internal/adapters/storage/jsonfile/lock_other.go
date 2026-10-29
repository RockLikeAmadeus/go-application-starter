//go:build !unix

package jsonfile

// lockDir is a no-op on platforms without flock. Single-process use is then
// the operator's responsibility.
func lockDir(string) (func() error, error) {
	return func() error { return nil }, nil
}
