// Package tui is a Bubble Tea driving adapter. It talks to the core only via
// driving ports, so it works identically in embedded and remote mode.
package tui

import (
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// Ports are the driving ports the TUI can show screens for.
type Ports struct {
	// sample:start
	Items ports.ItemService
	// sample:end
}

// Run starts the TUI and blocks until the user quits.
func Run(p Ports, opts ...tea.ProgramOption) error {
	_, err := tea.NewProgram(firstScreen(p), opts...).Run()
	return err
}

// firstScreen picks the screen the TUI opens on.
func firstScreen(p Ports) tea.Model {
	// sample:start
	if p.Items != nil {
		return New(p.Items)
	}
	// sample:end
	return home{}
}

// home is shown until the project has a capability of its own.
type home struct{}

func (home) Init() tea.Cmd { return nil }

func (h home) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyPressMsg); ok {
		switch k.String() {
		case "q", "esc", "ctrl+c":
			return h, tea.Quit
		}
	}
	return h, nil
}

func (home) View() tea.View {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Home") + "\n\n")
	b.WriteString(itemStyle.Render("Nothing here yet.") + "\n\n")
	b.WriteString(helpStyle.Render("q quit"))
	v := tea.NewView(frameStyle.Render(b.String()))
	v.AltScreen = true
	return v
}
