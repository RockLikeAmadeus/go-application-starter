package term_test

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/test/terminal/term"
)

func TestScreenShowsWhatTheProgramPrinted(t *testing.T) {
	tm, err := term.Start(exec.Command("sh", "-c", "printf hello"))
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = tm.Close() })

	deadline := time.Now().Add(5 * time.Second)
	for !strings.Contains(tm.Screen(), "hello") {
		if time.Now().After(deadline) {
			t.Fatalf("screen never showed %q; it shows:\n%s", "hello", tm.Screen())
		}
		time.Sleep(10 * time.Millisecond)
	}
}
