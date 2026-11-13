package browsersignin_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/browsersignin"
)

// A real browser must never open during tests, so BROWSER names a script that
// only records the address it was given.
func TestOpen_RunsTheProgramNamedByBrowserWithTheAddressAsItsLastArgument(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake browser is a shell script")
	}
	folder := t.TempDir()
	record := filepath.Join(folder, "recorded")
	script := filepath.Join(folder, "fakebrowser")
	body := "#!/bin/sh\necho \"$@\" > \"" + record + "\"\n"
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BROWSER", script+" --new-window")

	if err := browsersignin.Open("https://app.example.com/auth/native?port=1"); err != nil {
		t.Fatalf("Open: %v", err)
	}

	recorded, err := os.ReadFile(record)
	if err != nil {
		t.Fatalf("the program named by BROWSER did not run: %v", err)
	}
	if got := strings.TrimSpace(string(recorded)); got != "--new-window https://app.example.com/auth/native?port=1" {
		t.Errorf("the program got %q, want its own argument and then the address", got)
	}
}

// A browser that keeps running (a first Firefox window, say) must not hold up
// the sign-in, so Open returns while the program is still going.
func TestOpen_DoesNotWaitForABrowserThatKeepsRunning(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake browser is a shell script")
	}
	script := filepath.Join(t.TempDir(), "slowbrowser")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nsleep 5\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BROWSER", script)

	started := time.Now()
	if err := browsersignin.Open("https://app.example.com/auth/native?port=1"); err != nil {
		t.Fatalf("Open: %v", err)
	}
	if waited := time.Since(started); waited > 3*time.Second {
		t.Errorf("Open waited %v for a browser that keeps running", waited)
	}
}

// A program that fails at once is reported, and so is one that cannot start.
func TestOpen_ReportsAProgramThatFailsAtOnceOrCannotStart(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake browser is a shell script")
	}
	script := filepath.Join(t.TempDir(), "failingbrowser")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nexit 3\n"), 0o700); err != nil {
		t.Fatal(err)
	}
	for name, program := range map[string]string{
		"exits non-zero": script,
		"cannot start":   filepath.Join(t.TempDir(), "missing"),
	} {
		t.Setenv("BROWSER", program)
		if err := browsersignin.Open("https://app.example.com/"); err == nil {
			t.Errorf("%s: Open reported no error", name)
		}
	}
}
