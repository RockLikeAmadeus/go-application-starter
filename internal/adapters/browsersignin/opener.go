package browsersignin

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

// startupGrace is how long Open watches the browser program for an early
// failure. A program still running after it is taken to be a browser that
// is open and in use, so Open returns instead of holding up the sign-in.
const startupGrace = time.Second

// Open opens address in the user's browser: the program named by the BROWSER
// environment variable (with any arguments it carries), or the operating
// system's default opener when BROWSER is unset or blank, given the address
// as its last argument. It reports a program that cannot start or that exits
// with an error within startupGrace; it does not wait for one that keeps
// running. The program's output is discarded.
func Open(address string) error {
	fields := strings.Fields(os.Getenv("BROWSER"))
	if len(fields) == 0 {
		fields = systemOpener()
	}
	command := exec.Command(fields[0], append(fields[1:], address)...)
	if err := command.Start(); err != nil {
		return err
	}
	finished := make(chan error, 1)
	go func() { finished <- command.Wait() }()
	select {
	case err := <-finished:
		return err
	case <-time.After(startupGrace):
		return nil
	}
}

// systemOpener is the program that opens an address in the default browser.
func systemOpener() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{"open"}
	case "windows":
		return []string{"rundll32", "url.dll,FileProtocolHandler"}
	default:
		return []string{"xdg-open"}
	}
}
