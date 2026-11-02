package acceptance

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/cucumber/godog"
)

// registerReadinessSteps binds the steps of features/readiness.feature,
// which only the server driver runs.
func (w *world) registerReadinessSteps(sc *godog.ScenarioContext) {
	sc.Step(`^the server says it is ready$`, w.theServerSaysItIsReady)
	sc.Step(`^the server says it isn't ready, because its data is unavailable$`, w.theServerSaysItIsNotReady)
	sc.Step(`^the server's data folder is removed$`, w.theServersDataFolderIsRemoved)
	sc.Step(`^the server's data folder is put back$`, w.theServersDataFolderIsPutBack)
}

func (w *world) theServerSaysItIsReady() error {
	status, body, err := w.server.readiness()
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return fmt.Errorf("readiness = %d %q, want %d", status, body, http.StatusOK)
	}
	return nil
}

func (w *world) theServerSaysItIsNotReady() error {
	status, body, err := w.server.readiness()
	if err != nil {
		return err
	}
	if status != http.StatusServiceUnavailable || !strings.Contains(body, "storage unavailable") {
		return fmt.Errorf("readiness = %d %q, want %d saying storage unavailable", status, body, http.StatusServiceUnavailable)
	}
	return nil
}

func (w *world) theServersDataFolderIsRemoved() error {
	return os.RemoveAll(w.server.dataDir)
}

func (w *world) theServersDataFolderIsPutBack() error {
	return os.MkdirAll(w.server.dataDir, 0o755)
}
