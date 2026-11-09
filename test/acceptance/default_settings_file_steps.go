package acceptance

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cucumber/godog"
	"github.com/knadh/koanf/parsers/toml/v2"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// registerDefaultSettingsFileSteps binds the steps of
// features/default-settings-file.feature, which only the desktop and mobile
// drivers run: the settings file is on their computer.
func (w *world) registerDefaultSettingsFileSteps(sc *godog.ScenarioContext) {
	sc.Step(`^my settings file sets "([^"]*)" to (.+)$`, w.mySettingsFileSets)
	sc.Step(`^my settings file doesn't set "([^"]*)"$`, w.mySettingsFileDoesNotSet)
	sc.Step(`^the mobile app has created no settings file$`, w.theMobileAppHasCreatedNoSettingsFile)
	sc.Step(`^I delete my settings file$`, w.iDeleteMySettingsFile)
	sc.Step(`^I replace my settings file with one choosing the data folder "([^"]*)"$`, w.iReplaceMySettingsFileWithOneChoosingTheDataFolder)
	sc.Step(`^my settings file is as I left it$`, w.mySettingsFileIsAsILeftIt)
	sc.Step(`^I name a settings file of my own in the environment$`, w.iNameASettingsFileOfMyOwnInTheEnvironment)
	sc.Step(`^I name a settings file of my own with a command option$`, w.iNameASettingsFileOfMyOwnWithACommandOption)
	sc.Step(`^I have no settings file in my settings folder$`, w.iHaveNoSettingsFileInMySettingsFolder)
}

// iNameASettingsFileOfMyOwnInTheEnvironment writes a settings file outside
// the settings folder and names it with STARTER_CONFIG_FILE.
func (w *world) iNameASettingsFileOfMyOwnInTheEnvironment() error {
	file, err := w.writeASettingsFileOfMyOwn()
	if err != nil {
		return err
	}
	w.local.env["STARTER_CONFIG_FILE"] = file
	return nil
}

// iNameASettingsFileOfMyOwnWithACommandOption writes a settings file outside
// the settings folder and names it with --config on every command.
func (w *world) iNameASettingsFileOfMyOwnWithACommandOption() error {
	file, err := w.writeASettingsFileOfMyOwn()
	if err != nil {
		return err
	}
	w.commandLineOptions = append(w.commandLineOptions, "--config", file)
	return nil
}

// writeASettingsFileOfMyOwn writes a settings file outside the settings
// folder, choosing a data folder beside it, and returns its path.
func (w *world) writeASettingsFileOfMyOwn() (string, error) {
	if w.local == nil {
		return "", errNoLocalComputer
	}
	file := filepath.Join(w.local.dir, "my-settings.toml")
	text := fmt.Sprintf("data_dir = %q\n", filepath.Join(w.local.dir, "my-data"))
	if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
		return "", fmt.Errorf("write a settings file of my own: %w", err)
	}
	return file, nil
}

// iHaveNoSettingsFileInMySettingsFolder checks the command line worked and
// left no settings file in the settings folder.
func (w *world) iHaveNoSettingsFileInMySettingsFolder() error {
	if w.local == nil {
		return errNoLocalComputer
	}
	if w.lastErr != nil {
		return fmt.Errorf("the command line failed: %w", w.lastErr)
	}
	_, err := os.Stat(w.local.settingsFile())
	if err == nil {
		return fmt.Errorf("there's a settings file at %s, want none", w.local.settingsFile())
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("look for a settings file: %w", err)
	}
	return nil
}

// iReplaceMySettingsFileWithOneChoosingTheDataFolder writes a settings file
// of the user's own in place of the one the desktop app created, and keeps
// its text for "my settings file is as I left it".
func (w *world) iReplaceMySettingsFileWithOneChoosingTheDataFolder(name string) error {
	if w.local == nil {
		return errNoLocalComputer
	}
	text := fmt.Sprintf("data_dir = %q\n", filepath.Join(w.local.dir, name))
	if err := os.WriteFile(w.local.settingsFile(), []byte(text), 0o644); err != nil {
		return fmt.Errorf("replace the settings file: %w", err)
	}
	w.settingsFileLeft = text
	return nil
}

// mySettingsFileIsAsILeftIt checks the settings file still has exactly the
// text the user last wrote to it.
func (w *world) mySettingsFileIsAsILeftIt() error {
	if w.local == nil {
		return errNoLocalComputer
	}
	text, err := os.ReadFile(w.local.settingsFile())
	if err != nil {
		return fmt.Errorf("read the settings file: %w", err)
	}
	if string(text) != w.settingsFileLeft {
		return fmt.Errorf("the settings file is now:\n%s\nwant it as I left it:\n%s", text, w.settingsFileLeft)
	}
	return nil
}

// iDeleteMySettingsFile removes the settings file from this computer's
// settings folder.
func (w *world) iDeleteMySettingsFile() error {
	if w.local == nil {
		return errNoLocalComputer
	}
	if err := os.Remove(w.local.settingsFile()); err != nil {
		return fmt.Errorf("delete the settings file: %w", err)
	}
	return nil
}

// theMobileAppHasCreatedNoSettingsFile checks there's no settings file where
// the mobile app would look for one: the app's folder in its storage.
func (w *world) theMobileAppHasCreatedNoSettingsFile() error {
	if w.phoneStorage == "" {
		return errors.New("only the mobile driver runs the mobile app: tag the scenario @mobile")
	}
	settingsFile, err := config.LocalConfigFile(config.Env{
		Getenv:        func(string) string { return "" },
		UserConfigDir: func() (string, error) { return w.phoneStorage, nil },
	})
	if err != nil {
		return err
	}
	_, err = os.Stat(settingsFile)
	if err == nil {
		return fmt.Errorf("the mobile app left a settings file at %s, want none", settingsFile)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("look for a settings file: %w", err)
	}
	return nil
}

// mySettingsFileSets checks the settings file sets key to value: a quoted
// string, true or false, or "the desktop app's data folder".
func (w *world) mySettingsFileSets(key, value string) error {
	settings, err := w.readSettingsFile()
	if err != nil {
		return err
	}
	want, err := w.settingValue(value)
	if err != nil {
		return err
	}
	got, ok := settings[key]
	if !ok {
		return fmt.Errorf("the settings file doesn't set %q, want %v", key, want)
	}
	if got != want {
		return fmt.Errorf("the settings file sets %q to %#v, want %#v", key, got, want)
	}
	return nil
}

func (w *world) mySettingsFileDoesNotSet(key string) error {
	settings, err := w.readSettingsFile()
	if err != nil {
		return err
	}
	if got, ok := settings[key]; ok {
		return fmt.Errorf("the settings file sets %q to %#v, want it left out", key, got)
	}
	return nil
}

// readSettingsFile reads the settings file in the settings folder of this
// computer.
func (w *world) readSettingsFile() (map[string]any, error) {
	if w.local == nil {
		return nil, errNoLocalComputer
	}
	text, err := os.ReadFile(w.local.settingsFile())
	if err != nil {
		return nil, fmt.Errorf("read the settings file: %w", err)
	}
	return toml.Parser().Unmarshal(text)
}

// settingValue is the value a step names: a quoted string, true or false, or
// "the desktop app's data folder", which with nothing choosing another is
// the app's folder in the settings folder.
func (w *world) settingValue(value string) (any, error) {
	switch value {
	case "true", "false":
		return value == "true", nil
	case "the desktop app's data folder":
		return filepath.Dir(w.local.settingsFile()), nil
	}
	text, err := strconv.Unquote(value)
	if err != nil {
		return nil, fmt.Errorf("unknown setting value %s", value)
	}
	return text, nil
}
