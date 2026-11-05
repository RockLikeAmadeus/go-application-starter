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

// Run starts the TUI and blocks until the user quits. warnings show above
// whichever screen is open.
func Run(drivingPorts Ports, warnings []string, opts ...tea.ProgramOption) error {
	_, err := tea.NewProgram(firstScreen(drivingPorts, warnings), opts...).Run()
	return err
}

// firstScreen picks the screen the TUI opens on.
func firstScreen(drivingPorts Ports, warnings []string) tea.Model {
	var screen tea.Model = home{}
	// sample:start
	if drivingPorts.Items != nil {
		screen = newItemsScreen(drivingPorts.Items)
	}
	// sample:end
	return warnedScreen{screen: screen, warnings: warnings}
}

// warnedScreen shows the app's warnings, each on its own line, above screen, and
// passes everything else to it. It wraps whichever screen opens first, so the
// warnings stay above it after the sample is removed.
type warnedScreen struct {
	screen   tea.Model
	warnings []string
}

func (w warnedScreen) Init() tea.Cmd { return w.screen.Init() }

func (w warnedScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	w.screen, cmd = w.screen.Update(msg)
	return w, cmd
}

func (w warnedScreen) View() tea.View {
	v := w.screen.View()
	var b strings.Builder
	for _, m := range w.warnings {
		b.WriteString(warningLine(m) + "\n")
	}
	v.Content = b.String() + v.Content
	return v
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
