package tui

import (
	"context"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/clientsignin"
)

// sessionMsg carries what the server says about the current session.
type sessionMsg struct{ session clientsignin.Session }

// signInGate shows the signed-out screen, instead of screen, on a server that
// has sign-in on when no one is signed in, and passes everything to screen
// otherwise.
type signInGate struct {
	screen    tea.Model
	client    SignInClient
	serverURL string
	signedOut bool
}

func (g signInGate) Init() tea.Cmd {
	return func() tea.Msg {
		session, err := g.client.CurrentSession(context.Background())
		if err != nil {
			// Let the screen show its own error when it loads.
			return sessionMsg{}
		}
		return sessionMsg{session}
	}
}

func (g signInGate) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m, ok := msg.(sessionMsg); ok {
		g.signedOut = m.session.IsSignInOn && m.session.Name == ""
		if g.signedOut {
			return g, nil
		}
		return g, g.screen.Init()
	}
	if g.signedOut {
		if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "q" {
			return g, tea.Quit
		}
		return g, nil
	}
	var cmd tea.Cmd
	g.screen, cmd = g.screen.Update(msg)
	return g, cmd
}

func (g signInGate) View() tea.View {
	if !g.signedOut {
		return g.screen.View()
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("Signed out") + "\n\n")
	b.WriteString(itemStyle.Render("Not signed in to "+g.serverURL+".") + "\n")
	b.WriteString(itemStyle.Render("Sign in to see your data.") + "\n\n")
	b.WriteString(helpStyle.Render("l sign in · q quit"))
	v := tea.NewView(frameStyle.Render(b.String()))
	v.AltScreen = true
	return v
}
