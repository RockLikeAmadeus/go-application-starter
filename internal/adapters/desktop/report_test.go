package desktop_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
)

// A desktop app that can't start says why on stderr, and ends with a failing
// exit status.
func TestAnAppThatCannotStartPrintsItsErrorOnStderr(t *testing.T) {
	var errOut bytes.Buffer

	status := desktop.RunAndReport(func() error { return errors.New("the cause") }, &errOut)

	if status != 1 {
		t.Errorf("exit status = %d, want 1", status)
	}
	if got, want := errOut.String(), "Error: the cause\n"; got != want {
		t.Errorf("stderr = %q, want %q", got, want)
	}
}
