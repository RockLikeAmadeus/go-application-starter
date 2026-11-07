package desktop

import (
	"fmt"
	"io"
)

// RunAndReport calls run, which runs the app until it stops, and returns the
// exit status. An error that stopped it goes on errOut (stderr). It is here
// rather than in cmd/desktop so that it can be tested without Wails.
func RunAndReport(run func() error, errOut io.Writer) int {
	if err := run(); err != nil {
		fmt.Fprintln(errOut, "Error:", err)
		return 1
	}
	return 0
}
