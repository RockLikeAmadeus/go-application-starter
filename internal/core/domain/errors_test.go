package domain_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

func TestErrNotFound_Message(t *testing.T) {
	want := "The requested record could not be found."
	if got := domain.ErrNotFound.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}
