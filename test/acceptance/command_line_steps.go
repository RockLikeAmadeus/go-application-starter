package acceptance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/browsersignin"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/clientsignin"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// signInTimeout is how long a scenario waits for the command line to sign in.
const signInTimeout = 10 * time.Second

// commandLine is the command line the scenarios sign in from: the real
// browser sign-in and HTTP client, with a token store in memory.
type commandLine struct {
	*memoryTokenStore
	client       *clientsignin.Client
	entitlements ports.EntitlementsService
	// sample:start
	items ports.ItemService
	// sample:end
}

// memoryTokenStore keeps one session token, as the credential store would.
type memoryTokenStore struct{ token string }

func (s *memoryTokenStore) Save(token string) (string, error) {
	s.token = token
	return clientsignin.KeptInFile, nil
}

func (s *memoryTokenStore) Forget() error {
	s.token = ""
	return nil
}

func (s *memoryTokenStore) source(context.Context) (string, error) { return s.token, nil }

// sessionReader adapts httpclient's session to clientsignin's.
type sessionReader struct{ *httpclient.SessionService }

func (r sessionReader) CurrentSession(ctx context.Context) (clientsignin.Session, error) {
	session, err := r.SessionService.CurrentSession(ctx)
	return clientsignin.Session{IsSignInOn: session.IsSignInOn, Name: session.Name}, err
}

// browserSignIn signs in to one server through browsersignin.
type browserSignIn struct{ serverURL string }

func (b browserSignIn) BrowserSignIn(ctx context.Context, open func(address string) error) (string, error) {
	attempt, err := browsersignin.Start(ctx, b.serverURL, false, open)
	if err != nil {
		return "", err
	}
	defer attempt.Close()
	return attempt.Wait(ctx)
}

func newCommandLine(serverURL string) (*commandLine, error) {
	store := &memoryTokenStore{}
	session, err := httpclient.NewSessionService(serverURL, httpclient.WithTokenSource(store.source))
	if err != nil {
		return nil, err
	}
	entitlements, err := httpclient.NewEntitlementsService(serverURL, httpclient.WithTokenSource(store.source))
	if err != nil {
		return nil, err
	}
	c := &commandLine{
		memoryTokenStore: store,
		client:           clientsignin.New(sessionReader{session}, session, browserSignIn{serverURL}, store),
		entitlements:     entitlements,
	}
	// sample:start
	if c.items, err = httpclient.NewItemService(serverURL, httpclient.WithTokenSource(store.source)); err != nil {
		return nil, err
	}
	// sample:end
	return c, nil
}

// nativeSwapRecorder keeps the latest request the server got to swap a
// sign-in code for a session token, so a scenario can send it again.
type nativeSwapRecorder struct {
	mu   sync.Mutex
	body []byte
}

func (r *nativeSwapRecorder) latest() []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.body
}

// recording serves with next, keeping each request to swap a sign-in code.
func (r *nativeSwapRecorder) recording(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost && req.URL.Path == "/auth/native/token" {
			body, _ := io.ReadAll(req.Body)
			req.Body = io.NopCloser(bytes.NewReader(body))
			r.mu.Lock()
			r.body = body
			r.mu.Unlock()
		}
		next.ServeHTTP(w, req)
	})
}

func (w *world) registerCommandLineSignInSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I (?:have )?signed in from the command line as "([^"]*)"$`, w.iSignInFromTheCommandLine)
	sc.Step(`^I sign in from the command line as "([^"]*)"$`, w.iSignInFromTheCommandLine)
	sc.Step(`^the command line is signed in as "([^"]*)"$`, w.theCommandLineIsSignedInAs)
	sc.Step(`^the command line says my workspace's plan is "([^"]*)"$`, w.theCommandLineSaysThePlanIs)
	sc.Step(`^the command line tries to see my workspace's plan(?: without signing in)?$`,
		w.theCommandLineTriesToSeeThePlan)
	sc.Step(`^the command line tells me to sign in$`, w.iAmToldToSignIn)
	sc.Step(`^I sign out from the command line$`, w.iSignOutFromTheCommandLine)
	sc.Step(`^the command line keeps no session for the server$`, w.theCommandLineKeepsNoSession)
	sc.Step(`^I am (?:still )?signed in as "([^"]*)" in the browser$`, w.iAmSignedInAs)
	sc.Step(`^I sign in as "([^"]*)" in the browser$`, w.iSignInAs)
	sc.Step(`^someone tries to use the same sign-in code again$`, w.someoneUsesTheSameCodeAgain)
	sc.Step(`^they are not given a session$`, w.theyAreNotGivenASession)
	// sample:start
	sc.Step(`^I create an item named "([^"]*)" from the command line$`, w.iCreateAnItemFromTheCommandLine)
	// sample:end
}

// theCommandLine is the scenario's command line, made when first needed.
func (w *world) theCommandLine() (*commandLine, error) {
	if w.signIn == nil {
		return nil, errNoSignIn
	}
	if w.commandLine == nil {
		c, err := newCommandLine(w.apiURL)
		if err != nil {
			return nil, err
		}
		w.commandLine = c
	}
	return w.commandLine, nil
}

func (w *world) iSignInFromTheCommandLine(ctx context.Context, name string) error {
	c, err := w.theCommandLine()
	if err != nil {
		return err
	}
	browser, ok := w.signIn.(nativeSignInBrowser)
	if !ok {
		return errors.New("this driver's sign-in can't finish the command line's: tag the scenario @command-line-sign-in")
	}
	// The browser opens the address a moment after the command line asks, and
	// meets the command line's loopback port on its way back.
	browserErrors := make(chan error, 1)
	open := func(address string) error {
		go func() { browserErrors <- finishInTheBrowser(browser, name, address) }()
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, signInTimeout)
	defer cancel()
	if _, err := c.client.SignIn(ctx, open); err != nil {
		select {
		case browserErr := <-browserErrors:
			if browserErr != nil {
				return fmt.Errorf("%w (the browser: %w)", err, browserErr)
			}
		default:
		}
		return err
	}
	return nil
}

// finishInTheBrowser signs in at address as name, and returns to the command
// line's loopback port, as the browser would.
func finishInTheBrowser(browser nativeSignInBrowser, name, address string) error {
	callback, err := browser.finishNativeSignIn(name, address)
	if err != nil {
		return err
	}
	resp, err := http.Get(callback)
	if err != nil {
		return err
	}
	return resp.Body.Close()
}

func (w *world) theCommandLineIsSignedInAs(ctx context.Context, name string) error {
	c, err := w.theCommandLine()
	if err != nil {
		return err
	}
	session, err := c.client.CurrentSession(ctx)
	if err != nil {
		return err
	}
	if session.Name != name {
		return fmt.Errorf("the command line is signed in as %q, want %q", session.Name, name)
	}
	return nil
}

func (w *world) theCommandLineSaysThePlanIs(ctx context.Context, plan string) error {
	c, err := w.theCommandLine()
	if err != nil {
		return err
	}
	got, err := c.entitlements.CurrentEntitlements(ctx)
	if err != nil {
		return err
	}
	if got.Plan != plan {
		return fmt.Errorf("the command line says the plan is %q, want %q", got.Plan, plan)
	}
	return nil
}

func (w *world) theCommandLineTriesToSeeThePlan(ctx context.Context) error {
	c, err := w.theCommandLine()
	if err != nil {
		return err
	}
	_, w.lastErr = c.entitlements.CurrentEntitlements(ctx)
	return nil
}

func (w *world) iSignOutFromTheCommandLine(ctx context.Context) error {
	c, err := w.theCommandLine()
	if err != nil {
		return err
	}
	return c.client.SignOut(ctx)
}

func (w *world) theCommandLineKeepsNoSession() error {
	c, err := w.theCommandLine()
	if err != nil {
		return err
	}
	if c.token != "" {
		return errors.New("the command line still keeps a session token for the server")
	}
	return nil
}

func (w *world) someoneUsesTheSameCodeAgain(ctx context.Context) error {
	body := w.nativeSwaps.latest()
	if body == nil {
		return errors.New("no sign-in code was swapped for a session token")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, w.apiURL+"/auth/native/token", bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	page, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	w.codeReplay = answer{status: response.StatusCode, body: string(page)}
	return nil
}

func (w *world) theyAreNotGivenASession() error {
	if w.codeReplay.status != http.StatusBadRequest {
		return fmt.Errorf("the second use of the code answered status %d: %.80q, want %d",
			w.codeReplay.status, w.codeReplay.body, http.StatusBadRequest)
	}
	return nil
}

// sample:start

func (w *world) iCreateAnItemFromTheCommandLine(ctx context.Context, name string) error {
	c, err := w.theCommandLine()
	if err != nil {
		return err
	}
	_, err = c.items.CreateItem(ctx, name)
	return err
}

// sample:end
