package testsupport_test

import (
	"runtime"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

type skipRecorder struct{ skipped bool }

func (*skipRecorder) Helper()                { /* nothing to mark */ }
func (s *skipRecorder) Skipf(string, ...any) { s.skipped = true }

// Windows refuses to delete a file that is open, so a test that removes a data
// folder the app holds open can't be set up there. Elsewhere it runs.
func TestSkipWhereOpenFilesCannotBeRemovedSkipsOnlyOnWindows(t *testing.T) {
	rec := &skipRecorder{}

	testsupport.SkipWhereOpenFilesCannotBeRemoved(rec)

	if want := runtime.GOOS == "windows"; rec.skipped != want {
		t.Errorf("skipped = %v on %s, want %v", rec.skipped, runtime.GOOS, want)
	}
}
