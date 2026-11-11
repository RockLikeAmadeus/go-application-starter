//go:build unix

package terminal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"
)

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
	sc.Step(`^I export the items with the command line$`, j.iExportTheItemsWithTheCommandLine)
	sc.Step(`^the export lists "([^"]*)" and then "([^"]*)"$`, j.theExportListsAndThen)
	sc.Step(`^a plans file whose local plan doesn't have item export, which is (shown locked|hidden) when off$`, j.aPlansFileWithoutItemExport)
	sc.Step(`^I ask the command line for help$`, j.iAskTheCommandLineForHelp)
	sc.Step(`^item export is listed as disabled on your plan in the help$`, j.itemExportIsListedAsDisabled)
	sc.Step(`^item export is not listed in the help$`, j.itemExportIsNotListed)
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

// iExportTheItemsWithTheCommandLine runs the export; the next step checks
// how it went.
func (j *journey) iExportTheItemsWithTheCommandLine() error {
	return j.run("item", "export")
}

// theExportListsAndThen checks the export succeeded, printing the CSV's
// header and then first's and second's rows, in that order.
func (j *journey) theExportListsAndThen(first, second string) error {
	if err := j.succeeds(); err != nil {
		return err
	}
	screen := j.term.Screen()
	headerAt := strings.Index(screen, "name,createdAt")
	firstAt, secondAt := strings.Index(screen, first+","), strings.Index(screen, second+",")
	if headerAt < 0 || firstAt < headerAt || secondAt < firstAt {
		return fmt.Errorf("the export doesn't list the header, then %q, then %q; the screen shows:\n%s", first, second, screen)
	}
	return nil
}

// aPlansFileWithoutItemExport makes the user's plans file one where the local
// plan has no features, and item export is shown (locked or hidden) when off.
func (j *journey) aPlansFileWithoutItemExport(shown string) error {
	whenOff := map[string]string{"shown locked": "show-locked", "hidden": "hide"}[shown]
	plans := `{
	"newWorkspacePlan": "local",
	"features": {"item-export": {"name": "Item export", "whenOff": "` + whenOff + `"}},
	"plans": {"local": {"limits": {}, "features": []}}
}`
	path := filepath.Join(j.home, "plans.jsonc")
	if err := os.WriteFile(path, []byte(plans), 0o600); err != nil {
		return err
	}
	j.settings = append(j.settings, "STARTER_PLANS_FILE="+path)
	return nil
}

func (j *journey) iAskTheCommandLineForHelp() error {
	return j.run("item", "--help")
}

func (j *journey) itemExportIsListedAsDisabled() error {
	if err := j.term.WaitFor("Export the items as CSV (not included in your current plan)", timeout); err != nil {
		return err
	}
	return j.succeeds()
}

// itemExportIsNotListed waits for the help to end, so all of it is on the
// screen, and checks it doesn't list the export.
func (j *journey) itemExportIsNotListed() error {
	if err := j.succeeds(); err != nil {
		return err
	}
	if screen := j.term.Screen(); strings.Contains(screen, "Export the items") {
		return fmt.Errorf("the help lists item export; the screen shows:\n%s", screen)
	}
	return nil
}
