package tui

import (
	"context"
	"errors"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/clientsignin"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// sessionMsg carries what the server says about the current session.
type sessionMsg struct{ session clientsignin.Session }

// addressMsg carries the address the browser sign-in opened.
type addressMsg struct {
	address string
	// openFailed is true when the browser could not be opened.
	openFailed bool
}

// signOutDoneMsg says the sign-out ended, and how.
type signOutDoneMsg struct{ err error }

// signInDoneMsg says the browser sign-in ended, and how.
type signInDoneMsg struct {
	name string
	err  error
}

// failedMsg is what a screen's message implements when it reports that a
// request failed, so the gate can see a refusal for want of a session whichever
// screen made the request.
type failedMsg interface{ failure() error }

// signInGate shows the signed-out screen, instead of screen, on a server that
// has sign-in on when no one is signed in, and passes everything to screen
// otherwise.
type signInGate struct {
	screen    tea.Model
	client    SignInClient
	serverURL string
	// openBrowser opens the sign-in address in the user's browser.
	openBrowser func(address string) error
	signedOut   bool
	// waiting is true while a browser sign-in runs, with its address once
	// known, and cancel ends it.
	waiting bool
	address string
	// openFailed is true when the browser could not be opened.
	openFailed bool
	cancel     context.CancelFunc
	// err is why the last sign-in failed, shown on the signed-out screen.
	err error
	// notice is a note on the signed-out screen after a sign-out the server
	// did not confirm.
	notice string
	// signOutErr is why the last sign-out failed, shown on the signed-in
	// screen, where the user stays because the token is still stored.
	signOutErr error
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
	if f, ok := msg.(failedMsg); ok && g.name != "" && errors.Is(f.failure(), domain.ErrNoWorkspace) {
		g.signedOut, g.name, g.signOutErr = true, "", nil
		g.notice = "The server at " + g.serverURL + " refused this request, because there is no valid session for it: " +
			"you are not signed in, or the session token has expired or been ended. Press l to sign in."
		return g, nil
	}
	if m, ok := msg.(signInDoneMsg); ok && g.waiting {
		g.waiting = false
		if m.err != nil {
			g.err = m.err
			return g, nil
		}
		g.signedOut, g.name = false, m.name
		return g, g.screen.Init()
	}
	if g.waiting {
		return g.updateWaiting(msg)
	}
	if g.signedOut {
		if k, ok := msg.(tea.KeyPressMsg); ok {
			switch k.String() {
			case "q":
				return g, tea.Quit
			case "l":
				return g.startSignIn()
			}
		}
		return g, nil
	}
	if m, ok := msg.(signOutDoneMsg); ok {
		var notEnded *clientsignin.SessionNotEndedError
		switch {
		case m.err == nil:
			g.signedOut, g.name, g.notice = true, "", ""
		case errors.As(m.err, &notEnded):
			g.signedOut, g.name = true, ""
			g.notice = "Signed out of " + g.serverURL + " on this computer only: " +
				"the server did not confirm that it ended the session (" + notEnded.Cause.Error() + ")."
		default:
			g.signOutErr = m.err
		}
		return g, nil
	}
	if _, ok := msg.(tea.KeyPressMsg); ok {
		g.signOutErr = nil
	}
	if k, ok := msg.(tea.KeyPressMsg); ok && k.String() == "o" && g.canSignOut() {
		return g, func() tea.Msg { return signOutDoneMsg{g.client.SignOut(context.Background())} }
	}
	var cmd tea.Cmd
	g.screen, cmd = g.screen.Update(msg)
	return g, cmd
}

// canSignOut is true when someone is signed in and the screen shows the
// sign-out key, which it doesn't while a text input has focus.
func (g signInGate) canSignOut() bool {
	s, ok := g.screen.(signOutScreen)
	return g.name != "" && ok && !s.typing()
}

// startSignIn begins the browser sign-in, which runs until it finishes or is
// cancelled.
func (g signInGate) startSignIn() (tea.Model, tea.Cmd) {
	ctx, cancel := context.WithCancel(context.Background())
	g.waiting, g.address, g.openFailed, g.cancel, g.err = true, "", false, cancel, nil
	addresses := make(chan addressMsg, 1)
	// A browser that can't be opened doesn't end the sign-in: the user can
	// open the address by hand, so open reports success either way.
	open := func(address string) error {
		failed := g.openBrowser != nil && g.openBrowser(address) != nil
		addresses <- addressMsg{address, failed}
		return nil
	}
	waitForAddress := func() tea.Msg {
		select {
		case m := <-addresses:
			return m
		case <-ctx.Done():
			return nil
		}
	}
	signIn := func() tea.Msg {
		result, err := g.client.SignIn(ctx, open)
		return signInDoneMsg{result.Name, err}
	}
	return g, tea.Batch(waitForAddress, signIn)
}

// updateWaiting handles messages while the browser sign-in runs: Esc cancels it.
func (g signInGate) updateWaiting(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case addressMsg:
		g.address, g.openFailed = m.address, m.openFailed
	case tea.KeyPressMsg:
		if m.String() == "esc" {
			g.cancel()
			g.waiting = false
		}
	}
	return g, nil
}

func (g signInGate) View() tea.View {
	if g.waiting {
		return g.waitingView()
	}
	if !g.signedOut {
		return g.signedInView()
	}
	var b strings.Builder
	b.WriteString(titleStyle.Render("Signed out") + "\n\n")
	b.WriteString(itemStyle.Render("Not signed in to "+g.serverURL+".") + "\n")
	b.WriteString(itemStyle.Render("Sign in to see your data.") + "\n")
	if g.err != nil {
		b.WriteString(errorStyle.Render("Signing in failed: "+g.err.Error()+".") + "\n")
	}
	if g.notice != "" {
		b.WriteString(itemStyle.Render(g.notice) + "\n")
	}
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("l sign in • q quit"))
	v := tea.NewView(frameStyle.Render(b.String()))
	v.AltScreen = true
	return v
}

func (g signInGate) waitingView() tea.View {
	var b strings.Builder
	b.WriteString(titleStyle.Render("Signing in") + "\n\n")
	if g.openFailed {
		b.WriteString(itemStyle.Render("Could not open a browser. To sign in, open this address in a browser on this computer: "+
			g.address+". Waiting for you to finish signing in there.") + "\n")
	} else {
		b.WriteString(itemStyle.Render("Waiting for you to finish signing in, in your browser.") + "\n")
	}
	if g.address != "" && !g.openFailed {
		b.WriteString(itemStyle.Render("If no browser opened, open this address in a browser on this computer: "+g.address) + "\n")
	}
	b.WriteString("\n" + helpStyle.Render(helpLine("esc cancel")))
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
	head := itemStyle.Render("Signed in to "+g.serverURL+" as "+g.name+".") + "\n"
	if g.signOutErr != nil {
		head += errorStyle.Render("Signing out failed: "+g.signOutErr.Error()+".") + "\n"
	}
	v.Content = head + v.Content
	return v
}
