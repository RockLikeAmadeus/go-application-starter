// Package term runs a program in a pseudo-terminal and reads its screen as a
// user would see it, for the journeys in test/terminal.
package term

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/x/vt"
	"github.com/creack/pty"
)

// The terminal's size, in characters.
const (
	cols = 80
	rows = 24
)

// Terminal is a program running in an 80×24 pseudo-terminal.
type Terminal struct {
	pty *os.File

	mu  sync.Mutex // guards emu: vt.SafeEmulator doesn't guard String
	emu *vt.Emulator
}

// Start runs cmd in a new terminal.
func Start(cmd *exec.Cmd) (*Terminal, error) {
	f, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: cols, Rows: rows})
	if err != nil {
		return nil, err
	}
	t := &Terminal{pty: f, emu: vt.NewEmulator(cols, rows)}
	go t.showOutput()
	return t, nil
}

// showOutput draws what the program prints on the screen, until the terminal
// closes.
func (t *Terminal) showOutput() {
	buf := make([]byte, 4096)
	for {
		n, err := t.pty.Read(buf)
		t.mu.Lock()
		_, _ = t.emu.Write(buf[:n])
		t.mu.Unlock()
		if err != nil {
			return
		}
	}
}

// Screen returns the text the terminal shows now.
func (t *Terminal) Screen() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.emu.String()
}

// WaitFor returns once the screen shows text. If it doesn't within timeout,
// the error shows the screen as it was.
func (t *Terminal) WaitFor(text string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for !strings.Contains(t.Screen(), text) {
		if time.Now().After(deadline) {
			return fmt.Errorf("the screen didn't show %q within %s; it shows:\n%s", text, timeout, t.Screen())
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

// Close frees the terminal.
func (t *Terminal) Close() error {
	return t.pty.Close()
}
