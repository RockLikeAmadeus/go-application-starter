//go:build unix

package terminal

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
)

// registerTerminalAppSteps registers the steps of features/terminal-app.feature.
func (j *journey) registerTerminalAppSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I open the terminal app$`, j.iOpenTheTerminalApp)
	sc.Step(`^it says how to quit$`, j.itSaysHowToQuit)
	sc.Step(`^I quit the terminal app$`, j.iQuitTheTerminalApp)
	sc.Step(`^the terminal shows what it showed before$`, j.theTerminalShowsWhatItShowedBefore)
}

func (j *journey) iOpenTheTerminalApp() error {
	return j.run("tui")
}

func (j *journey) itSaysHowToQuit() error {
	return j.term.WaitFor("q quit", timeout)
}

func (j *journey) iQuitTheTerminalApp() error {
	return j.term.Press("q")
}

// theTerminalShowsWhatItShowedBefore checks the terminal app has ended and
// left the terminal as it found it: off the alternate screen, and as empty
// as each scenario's new terminal is.
func (j *journey) theTerminalShowsWhatItShowedBefore() error {
	if err := j.succeeds(); err != nil {
		return err
	}
	if j.term.AltScreen() {
		return fmt.Errorf("the terminal is still on the alternate screen; it shows:\n%s", j.term.Screen())
	}
	if screen := j.term.Screen(); strings.Contains(screen, "q quit") {
		return fmt.Errorf("the terminal app is still on the screen:\n%s", screen)
	}
	return nil
}
