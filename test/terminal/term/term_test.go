package term_test

import (
	"fmt"
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

func TestTypedTextAndEnterReachTheProgram(t *testing.T) {
	tm := start(t, `read line; printf 'got:%s' "$line"; sleep 5`)

	tm.Type("hi")
	if err := tm.Press("enter"); err != nil {
		t.Fatal(err)
	}

	if err := tm.WaitFor("got:hi", 2*time.Second); err != nil {
		t.Fatal(err)
	}
}

func TestPressSendsTheKeyAsATerminalWould(t *testing.T) {
	tests := []struct {
		key   string
		bytes string // as od prints them
	}{
		{"enter", "0d"},
		{"down", "1b 5b 42"},
		{"up", "1b 5b 41"},
		{"j", "6a"},
		{"k", "6b"},
		{"q", "71"},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			n := len(strings.Fields(tt.bytes))
			tm := start(t, fmt.Sprintf("stty raw -echo; printf ready; head -c %d | od -An -tx1; sleep 5", n))
			if err := tm.WaitFor("ready", 2*time.Second); err != nil {
				t.Fatal(err)
			}

			if err := tm.Press(tt.key); err != nil {
				t.Fatal(err)
			}

			if err := tm.WaitFor(tt.bytes, 2*time.Second); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPressFailsOnAKeyItDoesNotKnow(t *testing.T) {
	tm := start(t, "sleep 5")

	err := tm.Press("pgdn")
	if err == nil {
		t.Fatal(`Press("pgdn") returned no error`)
	}
	if !strings.Contains(err.Error(), `"pgdn"`) {
		t.Errorf("error %q doesn't name the key", err)
	}
}

func TestWaitReturnsTheExitStatus(t *testing.T) {
	tm := start(t, "exit 3")

	status, err := tm.Wait(2 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if status != 3 {
		t.Errorf("exit status = %d, want 3", status)
	}
}

func TestWaitFailsWithTheScreenWhenTheProgramKeepsRunning(t *testing.T) {
	tm := start(t, "printf hello; sleep 5")
	if err := tm.WaitFor("hello", 2*time.Second); err != nil {
		t.Fatal(err)
	}

	_, err := tm.Wait(300 * time.Millisecond)
	if err == nil {
		t.Fatal("Wait returned no error, but the program was still running")
	}
	if !strings.Contains(err.Error(), "hello") {
		t.Errorf("error %q doesn't show the screen", err)
	}
}

func TestAltScreenIsOnWhileTheProgramUsesIt(t *testing.T) {
	tm := start(t, `printf '\033[?1049h'; printf full-screen; sleep 5`)
	if err := tm.WaitFor("full-screen", 2*time.Second); err != nil {
		t.Fatal(err)
	}

	if !tm.AltScreen() {
		t.Error("AltScreen() = false while the program is on the alternate screen")
	}
}

func TestCloseStopsAProgramThatIgnoresHangUp(t *testing.T) {
	tm := start(t, `trap '' HUP; printf ready; sleep 30`)
	if err := tm.WaitFor("ready", 2*time.Second); err != nil {
		t.Fatal(err)
	}

	if err := tm.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := tm.Wait(2 * time.Second); err != nil {
		t.Fatalf("the program was still running after Close: %v", err)
	}
}

func TestWaitAgainAfterATimeoutReturnsTheExitStatus(t *testing.T) {
	tm := start(t, "read line; printf bye; exit 4")
	if _, err := tm.Wait(200 * time.Millisecond); err == nil {
		t.Fatal("the first Wait returned no error, but the program was waiting for a line")
	}

	if err := tm.Press("enter"); err != nil {
		t.Fatal(err)
	}
	if err := tm.WaitFor("bye", 2*time.Second); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond) // the program has exited, and the first Wait has seen it

	status, err := tm.Wait(2 * time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if status != 4 {
		t.Errorf("exit status = %d, want 4", status)
	}
}
