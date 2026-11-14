package tui

import (
	"context"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/clientsignin"
)

// fakeSignIn is a signed-in-or-not view of a server.
type fakeSignIn struct{ session clientsignin.Session }

func (f fakeSignIn) CurrentSession(context.Context) (clientsignin.Session, error) {
	return f.session, nil
}

// run plays the commands a model returns back into it, as Bubble Tea would,
// until no command is left to run.
func run(m tea.Model, cmd tea.Cmd) tea.Model {
	for cmd != nil {
		msg := cmd()
		if msg == nil {
			break
		}
		m, cmd = m.Update(msg)
	}
	return m
}

func TestOnAServerWithSignInOnAndNoSessionItShowsTheSignedOutScreenWithTheSignInKey(t *testing.T) {
	m := firstScreen(Ports{
		SignIn:    fakeSignIn{clientsignin.Session{IsSignInOn: true}},
		ServerURL: "https://app.example.com",
	}, nil)

	m = run(m, m.Init())

	got := m.View().Content
	for _, want := range []string{
		"Signed out",
		"Not signed in to https://app.example.com.",
		"Sign in to see your data.",
		"l sign in • q quit",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("view = %q, want it to contain %q", got, want)
		}
	}
}

func TestOnLocalDataOrWithSignInOffThereIsNoSignInOrSignOutKeyAndNoText(t *testing.T) {
	cases := map[string]Ports{
		"local data":  {},
		"sign-in off": {SignIn: fakeSignIn{clientsignin.Session{}}, ServerURL: "https://app.example.com"},
	}
	for name, ports := range cases {
		t.Run(name, func(t *testing.T) {
			m := firstScreen(ports, nil)
			m = run(m, m.Init())
			before := m.View().Content

			for _, key := range []rune{'l', 'o'} {
				m, _ = m.Update(tea.KeyPressMsg{Code: key, Text: string(key)})
			}

			got := m.View().Content
			for _, unwanted := range []string{"sign in", "Sign in", "sign out", "Signed"} {
				if strings.Contains(got, unwanted) || strings.Contains(before, unwanted) {
					t.Errorf("view contains %q: %q", unwanted, got)
				}
			}
		})
	}
}

func TestSignedInItShowsWhoIsSignedInAndTheSignOutKey(t *testing.T) {
	m := firstScreen(Ports{
		SignIn:    fakeSignIn{clientsignin.Session{IsSignInOn: true, Name: "Ada"}},
		ServerURL: "https://app.example.com",
	}, nil)

	m = run(m, m.Init())

	got := m.View().Content
	for _, want := range []string{
		"Signed in to https://app.example.com as Ada.",
		"o sign out • q quit",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("view = %q, want it to contain %q", got, want)
		}
	}
}
