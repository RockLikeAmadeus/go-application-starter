package tui

import (
	"errors"
	"testing"
)

func TestErrorLineShowsTheMessageInTheErrorStyle(t *testing.T) {
	got := errorLine(errors.New("boom"))

	if want := errorStyle.Render("Error: boom"); got != want {
		t.Errorf("errorLine = %q, want %q", got, want)
	}
}
