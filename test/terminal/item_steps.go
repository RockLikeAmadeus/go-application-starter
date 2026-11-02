package terminal

import "github.com/cucumber/godog"

// registerItemSteps registers the steps of features/items.feature, the
// sample's journeys.
func (j *journey) registerItemSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I add an item named "([^"]*)" with the command line$`, j.iAddAnItemNamedWithTheCommandLine)
	sc.Step(`^the command line lists "([^"]*)"$`, j.theCommandLineLists)
	sc.Step(`^I add an item named "([^"]*)" in the terminal app$`, j.iAddAnItemNamedInTheTerminalApp)
	sc.Step(`^the terminal app lists "([^"]*)"$`, j.theTerminalAppLists)
	sc.Step(`^I have added the items "([^"]*)" and "([^"]*)"$`, j.iHaveAddedTheItems)
	sc.Step(`^I press (\S+)$`, j.iPress)
	sc.Step(`^"([^"]*)" is selected$`, j.isSelected)
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

// iAddAnItemNamedInTheTerminalApp adds the item on the Items screen, and waits
// until the screen is back to browsing, so the name it then shows is the
// list's, not what was typed.
func (j *journey) iAddAnItemNamedInTheTerminalApp(name string) error {
	if err := j.term.WaitFor("a add", timeout); err != nil {
		return err
	}
	if err := j.term.Press("a"); err != nil {
		return err
	}
	if err := j.term.WaitFor("New item:", timeout); err != nil {
		return err
	}
	j.term.Type(name)
	if err := j.term.Press("enter"); err != nil {
		return err
	}
	return j.term.WaitFor("a add", timeout)
}

func (j *journey) theTerminalAppLists(name string) error {
	return j.term.WaitFor(name, timeout)
}

// iHaveAddedTheItems adds the items with the command line, before the
// terminal app opens.
func (j *journey) iHaveAddedTheItems(first, second string) error {
	for _, name := range []string{first, second} {
		if err := j.run("item", "add", name); err != nil {
			return err
		}
		if err := j.succeeds(); err != nil {
			return err
		}
	}
	return nil
}

// iPress presses key once a row is selected, so the list has loaded and the
// key can move through it.
func (j *journey) iPress(key string) error {
	if err := j.term.WaitFor(selectionMark, timeout); err != nil {
		return err
	}
	return j.term.Press(key)
}

func (j *journey) isSelected(name string) error {
	return j.term.WaitFor(selected(name), timeout)
}

// selectionMark marks the terminal app's selected row (selectedRow).
const selectionMark = "> "

// selected is how the terminal app shows name when its row is selected.
func selected(name string) string {
	return selectionMark + name
}
