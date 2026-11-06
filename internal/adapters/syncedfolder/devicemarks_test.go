package syncedfolder_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/syncedfolder"
)

var markTestNow = time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC)

// markFolder leaves device's mark in dir, as a sync service would bring it.
func markFolder(t *testing.T, dir, device string, markedAt time.Time) {
	t.Helper()
	data, err := json.Marshal(map[string]any{"device": device, "time": markedAt})
	if err != nil {
		t.Fatal(err)
	}
	devices := filepath.Join(dir, ".devices")
	if err := os.MkdirAll(devices, 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(devices, device+".json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestOtherDevicesReturnsDevicesWithARecentMark(t *testing.T) {
	dir := t.TempDir()
	markFolder(t, dir, "laptop", markTestNow.Add(-time.Minute))
	markFolder(t, dir, "phone", markTestNow.Add(-2*time.Minute))

	got := syncedfolder.OtherDevices(dir, "desktop", markTestNow)

	if want := []string{"laptop", "phone"}; !slices.Equal(got, want) {
		t.Errorf("OtherDevices() = %q, want %q", got, want)
	}
}
