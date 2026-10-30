package acceptance

import (
	"errors"
	"fmt"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
)

// registerDesktopSteps binds the steps of features/desktop.feature, which
// only the desktop driver runs.
func (w *world) registerDesktopSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I close the desktop app and open it again$`, w.iCloseTheDesktopAppAndOpenItAgain)
	sc.Step(`^I try to open a second copy of the desktop app on the same data folder$`, w.iTryToOpenASecondCopy)
	sc.Step(`^I am told the data is already in use$`, w.iAmToldTheDataIsInUse)
}

func (w *world) iCloseTheDesktopAppAndOpenItAgain() error {
	return w.reopen()
}

func (w *world) iTryToOpenASecondCopy() error {
	w.lastErr = w.openCopy()
	return nil
}

func (w *world) iAmToldTheDataIsInUse() error {
	if !errors.Is(w.lastErr, jsonfile.ErrLocked) {
		return fmt.Errorf("expected the data to be in use, got %w", w.lastErr)
	}
	return nil
}
