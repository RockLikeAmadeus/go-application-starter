package testsupport

import "runtime"

// Skipper is the part of *testing.T that SkipWhereOpenFilesCannotBeRemoved
// uses. Taking it as an interface keeps the testing package out of the
// binaries that import this one.
type Skipper interface {
	Helper()
	Skipf(format string, args ...any)
}

// SkipWhereOpenFilesCannotBeRemoved skips a test that removes a folder whose
// files the app holds open. Windows refuses to delete an open file, so the
// situation can't be set up there. Linux CI still proves the behaviour.
func SkipWhereOpenFilesCannotBeRemoved(t Skipper) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skipf("Windows refuses to delete a file the app holds open")
	}
}
