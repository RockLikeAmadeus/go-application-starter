package tui

import (
	"context"
	"errors"
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
	// signsInAs, when set, ends the sign-in at once, with this name.
	signsInAs string
	// failsWith, when set, ends the sign-in at once, with this error.
	failsWith error
	// signedOut counts the calls to SignOut, when set.
	signedOut *int
	// signOutErr, when set, is what SignOut returns.
	signOutErr error
}

// SignOut counts the call.
func (f fakeSignIn) SignOut(context.Context) error {
	if f.signedOut != nil {
		*f.signedOut++
	}
	return f.signOutErr
}

// SignIn opens the address, then waits for the sign-in to be cancelled.
func (f fakeSignIn) SignIn(ctx context.Context, open func(address string) error) (clientsignin.Result, error) {
	_ = open(f.address)
	if f.failsWith != nil {
		return clientsignin.Result{}, f.failsWith
	}
	if f.signsInAs != "" {
		return clientsignin.Result{Name: f.signsInAs}, nil
	}
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

func TestWhenTheBrowserCannotBeOpenedTheWaitingScreenSaysSoWithTheAddressAndKeepsWaiting(t *testing.T) {
	m := firstScreen(Ports{
		SignIn: fakeSignIn{
			session: clientsignin.Session{IsSignInOn: true},
			address: "https://app.example.com/auth/native?port=1",
		},
		ServerURL:   "https://app.example.com",
		OpenBrowser: func(string) error { return errors.New("no browser") },
	}, nil)
	m = run(m, m.Init())
	msgs := make(chan tea.Msg, 4)

	m, cmd := m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	collectAsync(cmd, msgs)
	m, _ = m.Update(next(t, msgs))

	got := m.View().Content
	for _, want := range []string{
		"Could not open a browser. To sign in, open this address in a browser on this computer: " +
			"https://app.example.com/auth/native?port=1. Waiting for you to finish signing in there.",
		"esc cancel",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("waiting view = %q, want it to contain %q", got, want)
		}
	}
	if strings.Contains(got, "Waiting for you to finish signing in, in your browser.") {
		t.Errorf("waiting view = %q, want the could-not-open text instead", got)
	}
	select {
	case msg := <-msgs:
		t.Errorf("got %T, want the sign-in to keep waiting", msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestAFinishedSignInShowsTheWorkspacesScreenOnTheNewSessionWithNoRestart(t *testing.T) {
	m := firstScreen(Ports{
		SignIn: fakeSignIn{
			session:   clientsignin.Session{IsSignInOn: true},
			address:   "https://app.example.com/auth/native?port=1",
			signsInAs: "Ada",
		},
		ServerURL:   "https://app.example.com",
		OpenBrowser: func(string) error { return nil },
	}, nil)
	m = run(m, m.Init())
	msgs := make(chan tea.Msg, 4)

	m, cmd := m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	collectAsync(cmd, msgs)
	for range 2 {
		var more tea.Cmd
		m, more = m.Update(next(t, msgs))
		collectAsync(more, msgs)
	}

	got := m.View().Content
	for _, want := range []string{
		"Signed in to https://app.example.com as Ada.",
		"Nothing here yet.",
		"o sign out • q quit",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("view = %q, want it to contain %q", got, want)
		}
	}
	for _, unwanted := range []string{"Signed out", "Signing in"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("view = %q, want no %q", got, unwanted)
		}
	}
}

func TestAFailedSignInShowsTheErrorOnTheSignedOutScreen(t *testing.T) {
	m := firstScreen(Ports{
		SignIn: fakeSignIn{
			session:   clientsignin.Session{IsSignInOn: true},
			address:   "https://app.example.com/auth/native?port=1",
			failsWith: errors.New("the server is unreachable"),
		},
		ServerURL:   "https://app.example.com",
		OpenBrowser: func(string) error { return nil },
	}, nil)
	m = run(m, m.Init())
	msgs := make(chan tea.Msg, 4)

	m, cmd := m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	collectAsync(cmd, msgs)
	for range 2 {
		var more tea.Cmd
		m, more = m.Update(next(t, msgs))
		collectAsync(more, msgs)
	}

	got := m.View().Content
	for _, want := range []string{
		"Signed out",
		"Signing in failed: the server is unreachable.",
		"l sign in • q quit",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("view = %q, want it to contain %q", got, want)
		}
	}
}

// failsOnce fails its first sign-in, then waits to be cancelled.
type failsOnce struct {
	fakeSignIn
	calls *int
}

func (f failsOnce) SignIn(ctx context.Context, open func(address string) error) (clientsignin.Result, error) {
	*f.calls++
	if *f.calls == 1 {
		_ = open(f.address)
		return clientsignin.Result{}, errors.New("the server is unreachable")
	}
	return f.fakeSignIn.SignIn(ctx, open)
}

func TestStartingANewSignInForgetsTheLastFailureSoCancellingShowsNoError(t *testing.T) {
	calls := 0
	m := firstScreen(Ports{
		SignIn: failsOnce{
			fakeSignIn: fakeSignIn{
				session: clientsignin.Session{IsSignInOn: true},
				address: "https://app.example.com/auth/native?port=1",
			},
			calls: &calls,
		},
		ServerURL:   "https://app.example.com",
		OpenBrowser: func(string) error { return nil },
	}, nil)
	m = run(m, m.Init())
	msgs := make(chan tea.Msg, 4)

	m, cmd := m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	collectAsync(cmd, msgs)
	for range 2 {
		var more tea.Cmd
		m, more = m.Update(next(t, msgs))
		collectAsync(more, msgs)
	}
	if got := m.View().Content; !strings.Contains(got, "Signing in failed") {
		t.Fatalf("view = %q, want the first failure shown", got)
	}

	m, cmd = m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	collectAsync(cmd, msgs)
	m, _ = m.Update(next(t, msgs))
	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

	got := m.View().Content
	if !strings.Contains(got, "Signed out") || strings.Contains(got, "Signing in failed") {
		t.Errorf("view after esc = %q, want the signed-out screen with no error", got)
	}
}

func TestTheSignOutKeySignsOutAndShowsTheSignedOutScreenWithNoMessage(t *testing.T) {
	var calls int
	m := firstScreen(Ports{
		SignIn:    fakeSignIn{session: clientsignin.Session{IsSignInOn: true, Name: "Ada"}, signedOut: &calls},
		ServerURL: "https://app.example.com",
	}, nil)
	m = run(m, m.Init())

	m, cmd := m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	m = run(m, cmd)

	if calls != 1 {
		t.Errorf("sign-outs = %d, want 1", calls)
	}
	got := m.View().Content
	for _, want := range []string{"Signed out", "Not signed in to https://app.example.com.", "l sign in • q quit"} {
		if !strings.Contains(got, want) {
			t.Errorf("view = %q, want it to contain %q", got, want)
		}
	}
	for _, unwanted := range []string{"Signed in to", "failed", "this computer only"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("view = %q, want no %q", got, unwanted)
		}
	}
}

func TestASignOutTheServerDidNotConfirmShowsTheSignedOutScreenWithTheNotice(t *testing.T) {
	m := firstScreen(Ports{
		SignIn: fakeSignIn{
			session:    clientsignin.Session{IsSignInOn: true, Name: "Ada"},
			signOutErr: &clientsignin.SessionNotEndedError{Cause: errors.New("server unreachable")},
		},
		ServerURL: "https://app.example.com",
	}, nil)
	m = run(m, m.Init())

	m, cmd := m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	m = run(m, cmd)

	got := m.View().Content
	for _, want := range []string{
		"Signed out",
		"Not signed in to https://app.example.com.",
		"Signed out of https://app.example.com on this computer only: " +
			"the server did not confirm that it ended the session (server unreachable).",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("view = %q, want it to contain %q", got, want)
		}
	}
}

// signOutsEndingAs is a fakeSignIn whose successive sign-outs end with each of
// results in turn.
type signOutsEndingAs struct {
	fakeSignIn
	results *[]error
}

func (f signOutsEndingAs) SignOut(context.Context) error {
	err := (*f.results)[0]
	*f.results = (*f.results)[1:]
	return err
}

func TestACleanSignOutShowsNoNoticeFromAnEarlierUnconfirmedOne(t *testing.T) {
	results := []error{&clientsignin.SessionNotEndedError{Cause: errors.New("server unreachable")}, nil}
	m := firstScreen(Ports{
		SignIn: signOutsEndingAs{
			fakeSignIn: fakeSignIn{session: clientsignin.Session{IsSignInOn: true, Name: "Ada"}, signsInAs: "Ada"},
			results:    &results,
		},
		ServerURL:   "https://app.example.com",
		OpenBrowser: func(string) error { return nil },
	}, nil)
	m = run(m, m.Init())
	press := func(key rune) {
		var cmd tea.Cmd
		m, cmd = m.Update(tea.KeyPressMsg{Code: key, Text: string(key)})
		m = run(m, cmd)
	}

	press('o')
	if got := m.View().Content; !strings.Contains(got, "this computer only") {
		t.Fatalf("view after the unconfirmed sign-out = %q, want the notice", got)
	}
	msgs := make(chan tea.Msg, 4)
	m, cmd := m.Update(tea.KeyPressMsg{Code: 'l', Text: "l"})
	collectAsync(cmd, msgs)
	for range 2 {
		var more tea.Cmd
		m, more = m.Update(next(t, msgs))
		collectAsync(more, msgs)
	}
	press('o')

	got := m.View().Content
	if !strings.Contains(got, "Signed out") || strings.Contains(got, "this computer only") {
		t.Errorf("view after the clean sign-out = %q, want the signed-out screen with no notice", got)
	}
}

func TestAFailedSignOutKeepsTheSignedInScreenAndShowsTheError(t *testing.T) {
	m := firstScreen(Ports{
		SignIn: fakeSignIn{
			session:    clientsignin.Session{IsSignInOn: true, Name: "Ada"},
			signOutErr: errors.New("keyring locked"),
		},
		ServerURL: "https://app.example.com",
	}, nil)
	m = run(m, m.Init())

	m, cmd := m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	m = run(m, cmd)

	got := m.View().Content
	for _, want := range []string{
		"Signed in to https://app.example.com as Ada.",
		"o sign out",
		"Signing out failed: keyring locked.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("view = %q, want it to contain %q", got, want)
		}
	}
	if strings.Contains(got, "Not signed in") {
		t.Errorf("view = %q, want the signed-in screen kept", got)
	}
}

func TestTheNextKeyPressClearsTheSignOutError(t *testing.T) {
	m := firstScreen(Ports{
		SignIn: fakeSignIn{
			session:    clientsignin.Session{IsSignInOn: true, Name: "Ada"},
			signOutErr: errors.New("keyring locked"),
		},
		ServerURL: "https://app.example.com",
	}, nil)
	m = run(m, m.Init())
	m, cmd := m.Update(tea.KeyPressMsg{Code: 'o', Text: "o"})
	m = run(m, cmd)

	m, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})

	if got := m.View().Content; strings.Contains(got, "Signing out failed") {
		t.Errorf("view = %q, want the error gone after a key press", got)
	}
}
