package acceptance

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

func TestSetSettingReplacesAKeyTheSettingsFileAlreadyHas(t *testing.T) {
	computer := &localComputer{dir: t.TempDir(), env: map[string]string{}}
	if err := os.MkdirAll(filepath.Dir(computer.settingsFile()), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(computer.settingsFile(), []byte("data_dir = \"old\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := computer.setSetting("data_dir", "new"); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(computer.configEnv(), computer.settingsFile())
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.DataDir != "new" {
		t.Errorf("data_dir = %q, want %q", cfg.DataDir, "new")
	}
}
