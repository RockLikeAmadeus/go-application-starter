package terminal

import "github.com/cucumber/godog"

// registerItemSteps registers the steps of features/items.feature, the
// sample's journeys.
func (j *journey) registerItemSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I add an item named "([^"]*)" with the command line$`, j.iAddAnItemNamedWithTheCommandLine)
	sc.Step(`^the command line lists "([^"]*)"$`, j.theCommandLineLists)
}

// iAddAnItemNamedWithTheCommandLine runs the command and waits for it to end,
// whatever its status: the next step checks how it went.
func (j *journey) iAddAnItemNamedWithTheCommandLine(name string) error {
	if err := j.run("item", "add", name); err != nil {
		return err
	}
	_, err := j.term.Wait(timeout)
	return err
}

func (j *journey) theCommandLineLists(name string) error {
	if err := j.run("item", "list"); err != nil {
		return err
	}
	if err := j.term.WaitFor(name, timeout); err != nil {
		return err
	}
	return j.succeeds()
}
