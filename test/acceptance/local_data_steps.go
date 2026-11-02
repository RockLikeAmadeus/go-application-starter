package acceptance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// registerLocalDataSteps binds the steps of features/local-data.feature,
// which only the desktop driver runs: the command line runs beside the
// desktop app, on the same computer.
func (w *world) registerLocalDataSteps(sc *godog.ScenarioContext) {
	sc.Step(`^I close the desktop app$`, w.iCloseTheDesktopApp)
	sc.Step(`^I ask the command line where my data is$`, w.iAskTheCommandLineWhereMyDataIs)
	sc.Step(`^I ask the command line where my data is, choosing the data folder "([^"]*)"$`, w.iAskChoosingTheDataFolder)
	sc.Step(`^it names the desktop app's data folder$`, w.itNamesTheDesktopAppsDataFolder)
	sc.Step(`^it names the data folder "([^"]*)"$`, w.itNamesTheDataFolder)
}

var errNoLocalComputer = errors.New("only the desktop driver runs the command line beside the desktop app: tag the scenario @desktop")

func (w *world) iCloseTheDesktopApp() error {
	if w.closeApp == nil {
		return errNoLocalComputer
	}
	w.closeApp()
	return nil
}

func (w *world) iAskTheCommandLineWhereMyDataIs(ctx context.Context) error {
	return w.runCommandLine(ctx, "info")
}

func (w *world) iAskChoosingTheDataFolder(ctx context.Context, name string) error {
	if w.local == nil {
		return errNoLocalComputer
	}
	return w.runCommandLine(ctx, "--data-dir", filepath.Join(w.local.dir, name), "info")
}

func (w *world) itNamesTheDesktopAppsDataFolder() error {
	if w.local == nil {
		return errNoLocalComputer
	}
	want, err := config.LocalDataDir(config.Env{Getenv: w.local.getenv, UserConfigDir: w.local.userConfigDir}, "")
	if err != nil {
		return err
	}
	return w.commandLineSaid(want)
}

func (w *world) itNamesTheDataFolder(name string) error {
	if w.local == nil {
		return errNoLocalComputer
	}
	return w.commandLineSaid(filepath.Join(w.local.dir, name))
}

// runCommandLine runs the CLI, as cmd/starter wires it, on the desktop app's
// computer. Its output and error are kept for the steps that follow.
func (w *world) runCommandLine(ctx context.Context, args ...string) error {
	if w.local == nil {
		return errNoLocalComputer
	}
	var out bytes.Buffer
	env := config.Env{Getenv: w.local.getenv, UserConfigDir: w.local.userConfigDir}
	cmd := cli.NewRootCmd(bootstrap.New, env, &out)
	cmd.SetArgs(args)
	w.lastErr = cmd.ExecuteContext(ctx)
	w.cliOut = out.String()
	return nil
}

// commandLineSaid checks that the command line succeeded and printed text.
func (w *world) commandLineSaid(text string) error {
	if w.lastErr != nil {
		return fmt.Errorf("the command line failed: %w", w.lastErr)
	}
	if !strings.Contains(w.cliOut, text) {
		return fmt.Errorf("the command line said:\n%s\nwant it to say %q", w.cliOut, text)
	}
	return nil
}
