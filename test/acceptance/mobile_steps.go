package acceptance

import "github.com/cucumber/godog"

// registerMobileSteps binds the steps of features/mobile.feature, which only
// the mobile driver runs.
func (w *world) registerMobileSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I close the mobile app and open it again$`, w.iCloseTheMobileAppAndOpenItAgain)
}

func (w *world) iCloseTheMobileAppAndOpenItAgain() error {
	return w.reopen()
}
