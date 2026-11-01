package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/memory"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/app"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
	"github.com/RockLikeAmadeus/go-application-starter/internal/testsupport"
)

// These tests drive the model directly: send messages, execute the returned
// commands synchronously, and assert on the rendered view.

func newCore() ports.ItemService {
	return app.NewItemService(memory.NewItemRepository(), testsupport.FixedWorkspace(domain.LocalWorkspace),
		testsupport.NewFixedClock(time.Unix(0, 0).UTC()), testsupport.NewSequentialIDs("item"))
}

// send applies msg and then feeds back the results of the model's own
// commands (item loads and errors). Other commands, such as cursor blinks,
// are ignored so tests stay fast and deterministic.
func send(t *testing.T, m tea.Model, msg tea.Msg) tea.Model {
	t.Helper()
	m, cmd := m.Update(msg)
	for _, next := range collect(cmd) {
		m = send(t, m, next)
	}
	return m
}

func collect(cmd tea.Cmd) []tea.Msg {
	if cmd == nil {
		return nil
	}
	done := make(chan tea.Msg, 1)
	go func() { done <- cmd() }()
	var msg tea.Msg
	select {
	case msg = <-done:
	case <-time.After(50 * time.Millisecond):
		return nil // a timer-based command; not ours
	}
	switch msg := msg.(type) {
	case tea.BatchMsg:
		var out []tea.Msg
		for _, c := range msg {
			out = append(out, collect(c)...)
		}
		return out
	case itemsLoadedMsg, errMsg:
		return []tea.Msg{msg}
	}
	return nil
}

func key(s string) tea.KeyPressMsg {
	switch s {
	case "enter":
		return tea.KeyPressMsg{Code: tea.KeyEnter}
	case "esc":
		return tea.KeyPressMsg{Code: tea.KeyEscape}
	}
	r := []rune(s)[0]
	return tea.KeyPressMsg{Code: r, Text: s}
}

func typeText(t *testing.T, m tea.Model, s string) tea.Model {
	for _, r := range s {
		m = send(t, m, key(string(r)))
	}
	return m
}

func view(m tea.Model) string { return m.View().Content }

func TestShowsEmptyState(t *testing.T) {
	m := tea.Model(New(newCore()))
	m = send(t, m, m.Init()())
	if !strings.Contains(view(m), "No items yet") {
		t.Errorf("view missing empty state:\n%s", view(m))
	}
}

func TestAddItem(t *testing.T) {
	core := newCore()
	m := tea.Model(New(core))
	m = send(t, m, m.Init()())

	m = send(t, m, key("a"))
	m = typeText(t, m, "Widget")
	m = send(t, m, key("enter"))

	if !strings.Contains(view(m), "Widget") {
		t.Errorf("view missing new item:\n%s", view(m))
	}
	items, _ := core.ListItems(context.Background())
	if len(items) != 1 || items[0].Name != "Widget" {
		t.Errorf("core items = %+v", items)
	}
}

func TestDeleteSelectedItem(t *testing.T) {
	core := newCore()
	_, _ = core.CreateItem(context.Background(), "Doomed")
	m := tea.Model(New(core))
	m = send(t, m, m.Init()())

	m = send(t, m, key("d"))

	if strings.Contains(view(m), "Doomed") {
		t.Errorf("item still shown after delete:\n%s", view(m))
	}
}

func TestValidationErrorIsShown(t *testing.T) {
	m := tea.Model(New(newCore()))
	m = send(t, m, m.Init()())

	m = send(t, m, key("a"))
	m = send(t, m, key("enter"))

	if !strings.Contains(view(m), "name must not be empty") {
		t.Errorf("view missing validation error:\n%s", view(m))
	}
}

func TestVimKeysMoveSelection(t *testing.T) {
	core := newCore()
	_, _ = core.CreateItem(context.Background(), "First")
	_, _ = core.CreateItem(context.Background(), "Second")
	m := tea.Model(New(core))
	m = send(t, m, m.Init()())

	m = send(t, m, key("j"))
	if !strings.Contains(view(m), "> Second") {
		t.Errorf("j did not select the next item:\n%s", view(m))
	}

	m = send(t, m, key("k"))
	if !strings.Contains(view(m), "> First") {
		t.Errorf("k did not select the previous item:\n%s", view(m))
	}
}

func TestVimKeysAreTypedWhileEditing(t *testing.T) {
	core := newCore()
	m := tea.Model(New(core))
	m = send(t, m, m.Init()())

	m = send(t, m, key("a"))
	m = typeText(t, m, "jk")
	send(t, m, key("enter"))

	items, _ := core.ListItems(context.Background())
	if len(items) != 1 || items[0].Name != "jk" {
		t.Errorf("core items = %+v", items)
	}
}
