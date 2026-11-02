package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

type mode int

const (
	modeBrowse mode = iota
	modeAdd
	modeRename
)

// Messages produced by commands.
type (
	itemsLoadedMsg struct{ items []domain.Item }
	errMsg         struct{ err error }
)

// Model is the Items screen.
type Model struct {
	items  ports.ItemService
	list   []domain.Item
	cursor int
	mode   mode
	input  textinput.Model
	err    error
}

// New creates the Items screen.
func New(items ports.ItemService) Model {
	in := textinput.New()
	in.Placeholder = "Item name"
	in.CharLimit = domain.MaxItemNameLength
	return Model{items: items, input: in}
}

func (m Model) Init() tea.Cmd {
	return m.load()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case itemsLoadedMsg:
		m.list = msg.items
		m.cursor = min(m.cursor, max(len(m.list)-1, 0))
		return m, nil
	case errMsg:
		m.err = msg.err
		return m, nil
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		if m.mode == modeBrowse {
			return m.updateBrowse(msg)
		}
		return m.updateInput(msg)
	}
	return m, nil
}

func (m Model) updateBrowse(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	m.err = nil
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		m.cursor = max(m.cursor-1, 0)
	case "down", "j":
		m.cursor = min(m.cursor+1, max(len(m.list)-1, 0))
	case "a":
		m.mode = modeAdd
		m.input.Reset()
		return m, m.input.Focus()
	case "r":
		if item, ok := m.selected(); ok {
			m.mode = modeRename
			m.input.SetValue(item.Name)
			return m, m.input.Focus()
		}
	case "d":
		if item, ok := m.selected(); ok {
			return m, m.run(func(ctx context.Context) error { return m.items.DeleteItem(ctx, item.ID) })
		}
	}
	return m, nil
}

func (m Model) updateInput(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = modeBrowse
		m.input.Blur()
		return m, nil
	case "enter":
		name := m.input.Value()
		mode := m.mode
		item, _ := m.selected()
		m.mode = modeBrowse
		m.input.Blur()
		return m, m.run(func(ctx context.Context) error {
			var err error
			if mode == modeAdd {
				_, err = m.items.CreateItem(ctx, name)
			} else {
				_, err = m.items.RenameItem(ctx, item.ID, name)
			}
			return err
		})
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Items") + "\n\n")

	if len(m.list) == 0 {
		b.WriteString(helpStyle.Render("  No items yet. Press a to add one.") + "\n")
	}
	for i, it := range m.list {
		if i == m.cursor && m.mode == modeBrowse {
			b.WriteString(selectedRow(it.Name) + "\n")
		} else {
			b.WriteString(itemStyle.Render(it.Name) + "\n")
		}
	}

	b.WriteString("\n")
	switch m.mode {
	case modeAdd:
		b.WriteString("New item: " + m.input.View() + "\n")
		b.WriteString(helpStyle.Render("enter save • esc cancel"))
	case modeRename:
		b.WriteString("Rename to: " + m.input.View() + "\n")
		b.WriteString(helpStyle.Render("enter save • esc cancel"))
	default:
		b.WriteString(helpStyle.Render("↑/↓ move • a add • r rename • d delete • q quit"))
	}
	if m.err != nil {
		b.WriteString("\n" + errorLine(m.err))
	}

	v := tea.NewView(frameStyle.Render(b.String()))
	v.AltScreen = true
	return v
}

func (m Model) selected() (domain.Item, bool) {
	if m.cursor < 0 || m.cursor >= len(m.list) {
		return domain.Item{}, false
	}
	return m.list[m.cursor], true
}

// load fetches the item list.
func (m Model) load() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		items, err := m.items.ListItems(ctx)
		if err != nil {
			return errMsg{fmt.Errorf("load items: %w", err)}
		}
		return itemsLoadedMsg{items}
	}
}

// run performs a mutation and then reloads the list.
func (m Model) run(op func(context.Context) error) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := op(ctx); err != nil {
			return errMsg{err}
		}
		return m.load()()
	}
}
