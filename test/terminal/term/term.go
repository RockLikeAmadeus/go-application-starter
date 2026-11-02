// Package term runs a program in a pseudo-terminal and reads its screen as a
// user would see it, for the journeys in test/terminal.
package term

import (
	"errors"
	"fmt"
	"io"
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
	cmd *exec.Cmd
	pty *os.File

	exited  chan struct{} // closed once the program has exited
	waitErr error         // cmd.Wait's result, set before exited closes

	mu  sync.Mutex // guards emu: vt.SafeEmulator doesn't guard String
	emu *vt.Emulator
}

// Start runs cmd in a new terminal.
func Start(cmd *exec.Cmd) (*Terminal, error) {
	f, err := pty.StartWithSize(cmd, &pty.Winsize{Cols: cols, Rows: rows})
	if err != nil {
		return nil, err
	}
	t := &Terminal{cmd: cmd, pty: f, exited: make(chan struct{}), emu: vt.NewEmulator(cols, rows)}
	go t.showOutput()
	go t.sendInput()
	go t.waitForExit()
	return t, nil
}

// waitForExit waits for the program once, for every Wait, since a command can
// be waited for only once.
func (t *Terminal) waitForExit() {
	t.waitErr = t.cmd.Wait()
	close(t.exited)
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

// sendInput sends the program what the terminal sends it: keys, and the
// terminal's answers to the program's questions (the emulator blocks until
// they're read), until the terminal closes.
func (t *Terminal) sendInput() {
	_, _ = io.Copy(t.pty, t.emu)
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
			return t.errorWithScreen(fmt.Sprintf("the screen didn't show %q within %s", text, timeout))
		}
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

// namedKeys are the keys Press knows by name.
var namedKeys = map[string]rune{
	"enter": vt.KeyEnter,
	"down":  vt.KeyDown,
	"up":    vt.KeyUp,
}

// Type sends text as if the user typed it.
func (t *Terminal) Type(text string) {
	t.emu.SendText(text)
}

// Press sends a key, as the terminal would encode it: one of namedKeys, or a
// single character such as "j" or "q".
func (t *Terminal) Press(key string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	code, ok := namedKeys[key]
	if !ok {
		chars := []rune(key)
		if len(chars) != 1 {
			return fmt.Errorf("unknown key %q: use one of namedKeys or a single character", key)
		}
		code = chars[0]
	}
	t.emu.SendKey(vt.KeyPressEvent{Code: code})
	return nil
}

// Wait waits for the program to exit and returns its exit status. If it
// doesn't exit within timeout, it returns an error.
func (t *Terminal) Wait(timeout time.Duration) (int, error) {
	select {
	case <-t.exited:
		var exit *exec.ExitError
		if errors.As(t.waitErr, &exit) {
			return exit.ExitCode(), nil
		}
		return 0, t.waitErr
	case <-time.After(timeout):
		return 0, t.errorWithScreen(fmt.Sprintf("the program didn't exit within %s", timeout))
	}
}

// AltScreen says whether the terminal is showing the alternate screen, the one
// full-screen programs draw on and leave when they exit.
func (t *Terminal) AltScreen() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.emu.IsAltScreen()
}

// errorWithScreen is an error saying what went wrong, then showing the screen,
// so a failed journey shows what the user would have seen.
func (t *Terminal) errorWithScreen(what string) error {
	return fmt.Errorf("%s; the screen shows:\n%s", what, t.Screen())
}

// Close stops the program, if it's still running, and frees the terminal.
func (t *Terminal) Close() error {
	_ = t.cmd.Process.Kill() // fails only if the program has already exited
	return t.pty.Close()
}
