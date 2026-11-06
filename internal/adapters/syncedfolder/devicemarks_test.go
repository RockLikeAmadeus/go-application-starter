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

func TestOtherDevicesLeavesOutThisDevice(t *testing.T) {
	dir := t.TempDir()
	markFolder(t, dir, "desktop", markTestNow.Add(-time.Minute))
	markFolder(t, dir, "laptop", markTestNow.Add(-time.Minute))

	got := syncedfolder.OtherDevices(dir, "desktop", markTestNow)

	if want := []string{"laptop"}; !slices.Equal(got, want) {
		t.Errorf("OtherDevices() = %q, want %q", got, want)
	}
}

func TestOtherDevicesLeavesOutMarksFiveMinutesOldOrMore(t *testing.T) {
	dir := t.TempDir()
	markFolder(t, dir, "laptop", markTestNow.Add(-5*time.Minute+time.Second))
	markFolder(t, dir, "phone", markTestNow.Add(-5*time.Minute))
	markFolder(t, dir, "tablet", markTestNow.Add(-time.Hour))

	got := syncedfolder.OtherDevices(dir, "desktop", markTestNow)

	if want := []string{"laptop"}; !slices.Equal(got, want) {
		t.Errorf("OtherDevices() = %q, want %q", got, want)
	}
}

func TestOtherDevicesLeavesOutFilesThatAreNotMarks(t *testing.T) {
	recent := markTestNow.Add(-time.Minute).Format(time.RFC3339)
	tests := []struct {
		name     string
		contents string
	}{
		{"not JSON", "not a mark"},
		{"empty", ""},
		{"no device", `{"time":"` + recent + `"}`},
		{"device isn't text", `{"device":5,"time":"` + recent + `"}`},
		{"time isn't a time", `{"device":"phone","time":"soon"}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			markFolder(t, dir, "laptop", markTestNow.Add(-time.Minute))
			notAMark := filepath.Join(dir, ".devices", "phone.json")
			if err := os.WriteFile(notAMark, []byte(tt.contents), 0o600); err != nil {
				t.Fatal(err)
			}

			got := syncedfolder.OtherDevices(dir, "desktop", markTestNow)

			if want := []string{"laptop"}; !slices.Equal(got, want) {
				t.Errorf("OtherDevices() = %q, want %q", got, want)
			}
		})
	}
}

func TestOtherDevicesOfAFolderWithNoMarksIsEmpty(t *testing.T) {
	got := syncedfolder.OtherDevices(t.TempDir(), "desktop", markTestNow)

	if len(got) != 0 {
		t.Errorf("OtherDevices() = %q, want none", got)
	}
}

func TestWriteMarkLeavesTheDevicesMarkInItsOwnFile(t *testing.T) {
	dir := t.TempDir()

	if err := syncedfolder.WriteMark(dir, "desktop", markTestNow); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".devices", "desktop.json"))
	if err != nil {
		t.Fatal(err)
	}
	var mark struct {
		Device string    `json:"device"`
		Time   time.Time `json:"time"`
	}
	if err := json.Unmarshal(data, &mark); err != nil {
		t.Fatal(err)
	}
	if mark.Device != "desktop" || !mark.Time.Equal(markTestNow) {
		t.Errorf("mark = %+v, want desktop at %v", mark, markTestNow)
	}
}

func TestWriteMarkReplacesTheOldFileRatherThanWritingOverIt(t *testing.T) {
	dir := t.TempDir()
	if err := syncedfolder.WriteMark(dir, "desktop", markTestNow); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, ".devices", "desktop.json")
	old, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := syncedfolder.WriteMark(dir, "desktop", markTestNow.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}

	replaced, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if os.SameFile(old, replaced) {
		t.Error("the mark was written over in place, so a sync service could copy half of it")
	}
}
