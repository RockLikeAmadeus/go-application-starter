//go:build unix

package terminal

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cucumber/godog"
)

// browserScript stands in for the user's browser, so no real one opens. The
// command line runs it with the sign-in address. It reads the name to sign in as
// from the file "browser-name" beside it, goes to the address, types the name
// into the dev sign-in's form, and follows the redirects back to the command
// line's loopback address, as the browser would.
const browserScript = `#!/bin/sh
dir=$(dirname "$0")
name=$(cat "$dir/browser-name") || exit 1
origin=$(echo "$1" | sed -E 's|^(https?://[^/]+)/.*|\1|')
curl -s -L -o /dev/null -c "$dir/browser-jar" -b "$dir/browser-jar" "$1"
curl -s -L -o /dev/null -c "$dir/browser-jar" -b "$dir/browser-jar" \
	--data-urlencode "name=$name" --data-urlencode "return=/auth/native/finish" "$origin/auth/dev"
`

// registerSignInSteps registers the steps of features/command-line-sign-in.feature.
func (j *journey) registerSignInSteps(sc *godog.ScenarioContext) {
	sc.Step(`^a server with the dev sign-in is running$`, j.aServerWithTheDevSignInIsRunning)
	sc.Step(`^a server is running with sign-in off$`, j.aServerIsRunning)
	sc.Step(`^I sign in to the server from the command line as "([^"]*)"$`, j.iSignInToTheServerFromTheCommandLineAs)
	sc.Step(`^I try to sign in to the server from the command line$`, j.iTryToSignInToTheServerFromTheCommandLine)
	sc.Step(`^the command line says I am signed in as "([^"]*)"$`, j.theCommandLineSaysIAmSignedInAs)
	sc.Step(`^the command line says I am signed in as "([^"]*)" on the server$`,
		j.theCommandLineSaysIAmSignedInAsOnTheServer)
	sc.Step(`^the command line says I am not signed in on the server$`, j.theCommandLineSaysIAmNotSignedInOnTheServer)
	sc.Step(`^it fails, saying the server doesn't sign people in$`, j.itFailsSayingTheServerDoesntSignPeopleIn)
	sc.Step(`^I (?:open|have opened) the terminal app, connected to the server, without signing in$`,
		j.iOpenTheTerminalAppConnectedToTheServer)
	sc.Step(`^the terminal app offers to sign me in$`, j.theTerminalAppOffersToSignMeIn)
	sc.Step(`^I sign in from the terminal app as "([^"]*)"$`, j.iSignInFromTheTerminalAppAs)
	sc.Step(`^I have signed in from the terminal app as "([^"]*)"$`, j.iHaveSignedInFromTheTerminalAppAs)
	sc.Step(`^the terminal app says I am signed in as "([^"]*)"$`, j.theTerminalAppSaysIAmSignedInAs)
	sc.Step(`^I sign out from the terminal app$`, j.iSignOutFromTheTerminalApp)
}

// installBrowser writes the stand-in browser into this scenario's computer and
// names it as the user's browser, and keeps the command line's token in a file
// instead of the developer's own credential store.
func (j *journey) installBrowser() error {
	script := filepath.Join(j.home, "browser.sh")
	if err := os.WriteFile(script, []byte(browserScript), 0o755); err != nil {
		return err
	}
	j.settings = append(j.settings, "BROWSER="+script, "STARTER_CREDENTIAL_STORE=file")
	return nil
}

// signInAs sets the name the stand-in browser signs in as.
func (j *journey) signInAs(name string) error {
	return os.WriteFile(filepath.Join(j.home, "browser-name"), []byte(name), 0o644)
}

func (j *journey) aServerWithTheDevSignInIsRunning() error {
	return j.startServerWith("STARTER_SIGN_IN=dev")
}

func (j *journey) iSignInToTheServerFromTheCommandLineAs(name string) error {
	if err := j.signInAs(name); err != nil {
		return err
	}
	return j.run("--server", j.serverURL, "login")
}

func (j *journey) iTryToSignInToTheServerFromTheCommandLine() error {
	return j.run("--server", j.serverURL, "login")
}

func (j *journey) theCommandLineSaysIAmSignedInAs(name string) error {
	if err := j.term.WaitFor(signedInLine(j.serverURL, name), timeout); err != nil {
		return err
	}
	return j.succeeds()
}

// theCommandLineSaysIAmSignedInAsOnTheServer asks the command line again, so the
// token must have been kept.
func (j *journey) theCommandLineSaysIAmSignedInAsOnTheServer(name string) error {
	if err := j.run("--server", j.serverURL, "info"); err != nil {
		return err
	}
	return j.theCommandLineSaysIAmSignedInAs(name)
}

func (j *journey) theCommandLineSaysIAmNotSignedInOnTheServer() error {
	if err := j.run("--server", j.serverURL, "info"); err != nil {
		return err
	}
	if err := j.term.WaitFor("Not signed in to "+j.serverURL, timeout); err != nil {
		return err
	}
	return j.succeeds()
}

func (j *journey) itFailsSayingTheServerDoesntSignPeopleIn() error {
	return j.itFailsSaying("does not support sign-in")
}

func (j *journey) iOpenTheTerminalAppConnectedToTheServer() error {
	return j.run("--server", j.serverURL, "tui")
}

func (j *journey) theTerminalAppOffersToSignMeIn() error {
	return j.term.WaitFor("l sign in", timeout)
}

// iSignInFromTheTerminalAppAs presses the sign-in key once the app offers it.
func (j *journey) iSignInFromTheTerminalAppAs(name string) error {
	if err := j.signInAs(name); err != nil {
		return err
	}
	if err := j.theTerminalAppOffersToSignMeIn(); err != nil {
		return err
	}
	return j.term.Press("l")
}

func (j *journey) iHaveSignedInFromTheTerminalAppAs(name string) error {
	if err := j.iOpenTheTerminalAppConnectedToTheServer(); err != nil {
		return err
	}
	if err := j.iSignInFromTheTerminalAppAs(name); err != nil {
		return err
	}
	return j.theTerminalAppSaysIAmSignedInAs(name)
}

func (j *journey) theTerminalAppSaysIAmSignedInAs(name string) error {
	return j.term.WaitFor(signedInLine(j.serverURL, name), timeout)
}

func (j *journey) iSignOutFromTheTerminalApp() error {
	return j.term.Press("o")
}

// signedInLine is what the command line and the terminal app print for who is
// signed in.
func signedInLine(serverURL, name string) string {
	return fmt.Sprintf("Signed in to %s as %s.", serverURL, name)
}
