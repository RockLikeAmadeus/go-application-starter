package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/clientsignin"
)

// fakeSignIn is a signed-in-or-not view of a server.
type fakeSignIn struct {
	session clientsignin.Session
	// address is what the sign-in hands to its opener before it waits.
	address string
}

// SignIn opens the address, then waits for the sign-in to be cancelled.
func (f fakeSignIn) SignIn(ctx context.Context, open func(address string) error) (clientsignin.Result, error) {
	_ = open(f.address)
	<-ctx.Done()
	return clientsignin.Result{}, ctx.Err()
}

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
		SignIn:    fakeSignIn{session: clientsignin.Session{IsSignInOn: true}},
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
		"sign-in off": {SignIn: fakeSignIn{session: clientsignin.Session{}}, ServerURL: "https://app.example.com"},
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
		SignIn:    fakeSignIn{session: clientsignin.Session{IsSignInOn: true, Name: "Ada"}},
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

// collectAsync runs cmd, and every command inside a batch, each in its own
// goroutine as Bubble Tea would, and sends their messages on the channel.
func collectAsync(cmd tea.Cmd, out chan<- tea.Msg) {
	if cmd == nil {
		return
	}
	go func() {
		msg := cmd()
		if batch, ok := msg.(tea.BatchMsg); ok {
			for _, c := range batch {
				collectAsync(c, out)
			}
			return
		}
		out <- msg
	}()
}

// next receives the next message, failing the test when none comes.
func next(t *testing.T, msgs <-chan tea.Msg) tea.Msg {
	t.Helper()
	select {
	case msg := <-msgs:
		return msg
	case <-time.After(2 * time.Second):
		t.Fatal("no message came")
		return nil
	}
}

func TestTheSignInKeyShowsTheWaitingScreenWithTheAddressAndEscReturnsToTheSignedOutScreen(t *testing.T) {
	var opened []string
	m := firstScreen(Ports{
		SignIn: fakeSignIn{
			session: clientsignin.Session{IsSignInOn: true},
			address: "https://app.example.com/auth/native?port=1",
		},
		ServerURL:   "https://app.example.com",
		OpenBrowser: func(address string) error { opened = append(opened, address); return nil },
	}, nil)
	m = run(m, m.Init())
	msgs := make(chan tea.Msg, 4)

	m, cmd := m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	collectAsync(cmd, msgs)
	m, _ = m.Update(next(t, msgs))

	got := m.View().Content
	for _, want := range []string{
		"Signing in",
		"Waiting for you to finish signing in, in your browser.",
		"If no browser opened, open this address in a browser on this computer: https://app.example.com/auth/native?port=1",
		"esc cancel",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("waiting view = %q, want it to contain %q", got, want)
		}
	}
	if len(opened) != 1 || opened[0] != "https://app.example.com/auth/native?port=1" {
		t.Errorf("opened = %v, want the address once", opened)
	}

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	m, _ = m.Update(next(t, msgs))

	got = m.View().Content
	for _, want := range []string{"Signed out", "l sign in • q quit"} {
		if !strings.Contains(got, want) {
			t.Errorf("view after esc = %q, want it to contain %q", got, want)
		}
	}
	if strings.Contains(got, "Signing in failed") {
		t.Errorf("view after esc = %q, want no error", got)
	}
}
