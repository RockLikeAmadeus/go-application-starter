package tui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestWithoutCapabilitiesShowsAPlaceholderThatQuits(t *testing.T) {
	m := firstScreen(Ports{}, nil)

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

func TestWarningsShowAboveWhicheverScreenIsOpen(t *testing.T) {
	m := firstScreen(Ports{}, []string{"first warning", "second warning"})

	assertWarnedAbove := func(m tea.Model, when string) {
		t.Helper()
		got := m.View().Content
		want := warningLine("first warning") + "\n" + warningLine("second warning")
		at, screen := strings.Index(got, want), strings.Index(got, "Home")
		if at < 0 || screen < 0 || at > screen {
			t.Errorf("%s, view = %q, want %q above the screen's title", when, got, want)
		}
	}
	assertWarnedAbove(m, "on opening")
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	assertWarnedAbove(m, "after an update")
}
