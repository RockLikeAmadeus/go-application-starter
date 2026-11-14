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
	// name is who is signed in, empty when no one is or sign-in is off.
	name string
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
		g.name = m.session.Name
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
		return g.signedInView()
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("Signed out") + "\n\n")
	b.WriteString(itemStyle.Render("Not signed in to "+g.serverURL+".") + "\n")
	b.WriteString(itemStyle.Render("Sign in to see your data.") + "\n\n")
	b.WriteString(helpStyle.Render("l sign in • q quit"))
	v := tea.NewView(frameStyle.Render(b.String()))
	v.AltScreen = true
	return v
}

// signOutBinding is the sign-out key's entry in a help line.
const signOutBinding = "o sign out"

// signOutScreen is what a screen implements to take part in the signed-in
// view. The gate never reads a screen's rendering.
type signOutScreen interface {
	tea.Model
	// typing reports whether a text input has focus, so letter keys such as
	// the sign-out key must reach the input instead.
	typing() bool
	// withExtraHelp returns the screen with binding added to its help line,
	// just before its quit key.
	withExtraHelp(binding string) tea.Model
}

// helpLine joins key bindings, skipping empty ones, as every help line does.
func helpLine(bindings ...string) string {
	var kept []string
	for _, b := range bindings {
		if b != "" {
			kept = append(kept, b)
		}
	}
	return strings.Join(kept, " • ")
}

// signedInView is the screen's view, with who is signed in above it and the
// sign-out key in its help line, when the screen takes part (see
// signOutScreen) and no text input has focus.
func (g signInGate) signedInView() tea.View {
	s, ok := g.screen.(signOutScreen)
	if g.name == "" || !ok || s.typing() {
		return g.screen.View()
	}
	v := s.withExtraHelp(signOutBinding).View()
	v.Content = itemStyle.Render("Signed in to "+g.serverURL+" as "+g.name+".") + "\n" + v.Content
	return v
}
