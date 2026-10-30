package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestWithoutCapabilitiesShowsAPlaceholderThatQuits(t *testing.T) {
	m := firstScreen(Ports{})

	if got := m.View().Content; !strings.Contains(got, "Nothing here yet") {
		t.Errorf("view = %q, want a placeholder saying there is nothing here yet", got)
	}
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
	if cmd == nil {
		t.Fatal("q returned no command, want tea.Quit")
	}
	if _, ok := cmd().(tea.QuitMsg); !ok {
		t.Errorf("q returned %T, want tea.QuitMsg", cmd())
	}
}
