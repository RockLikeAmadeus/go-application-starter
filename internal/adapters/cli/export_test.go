package cli

import (
	tea "charm.land/bubbletea/v2"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/tui"
)

// RunTUIWith makes `starter tui` call run instead of starting the terminal UI,
// and returns a function that puts the real one back.
func RunTUIWith(run func(tui.Ports, []string, ...tea.ProgramOption) error) (restore func()) {
	original := runTUI
	runTUI = run
	return func() { runTUI = original }
}
