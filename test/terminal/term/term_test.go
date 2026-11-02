package term_test

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/test/terminal/term"
)

// start runs a shell script in a new terminal, closed when the test ends.
func start(t *testing.T, script string) *term.Terminal {
	t.Helper()
	tm, err := term.Start(exec.Command("sh", "-c", script))
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { _ = tm.Close() })
	return tm
}

func TestScreenShowsWhatTheProgramPrinted(t *testing.T) {
	tm := start(t, "printf hello")

	if err := tm.WaitFor("hello", 5*time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestWaitForFailsWithTheScreenWhenTheTextNeverShows(t *testing.T) {
	tm := start(t, "printf hello; sleep 5")

	err := tm.WaitFor("goodbye", 300*time.Millisecond)
	if err == nil {
		t.Fatal("WaitFor returned no error, but the screen never showed \"goodbye\"")
	}
	for _, want := range []string{`"goodbye"`, "hello"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q doesn't contain %s", err, want)
		}
	}
}
