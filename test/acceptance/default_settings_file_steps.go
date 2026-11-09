package acceptance

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cucumber/godog"
	"github.com/knadh/koanf/parsers/toml/v2"
)

// registerDefaultSettingsFileSteps binds the steps of
// features/default-settings-file.feature, which only the desktop and mobile
// drivers run: the settings file is on their computer.
func (w *world) registerDefaultSettingsFileSteps(sc *godog.ScenarioContext) {
	sc.Step(`^my settings file sets "([^"]*)" to (.+)$`, w.mySettingsFileSets)
	sc.Step(`^my settings file doesn't set "([^"]*)"$`, w.mySettingsFileDoesNotSet)
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
