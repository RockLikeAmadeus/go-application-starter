package syncedfolder

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// deviceMarksFolder holds a file for each device that marked the folder.
const deviceMarksFolder = ".devices"

// markStaleAfter is the age from which a mark no longer counts.
const markStaleAfter = 5 * time.Minute

// A deviceMark is the file a device leaves to say it has the folder in use.
type deviceMark struct {
	Device string    `json:"device"`
	Time   time.Time `json:"time"`
}

// OtherDevices returns the devices other than thisDevice with a recent mark in
// dir, as of now. A mark's time comes from the device that wrote it, so a
// skewed clock can miss a device or keep a stale one. Files that aren't marks
// are left out, and a folder with no marks, or none that can be read, has no
// other devices: this is a best guess, so it never stops the app.
func OtherDevices(dir, thisDevice string, now time.Time) []string {
	marksDir := filepath.Join(dir, deviceMarksFolder)
	entries, _ := os.ReadDir(marksDir)
	var devices []string
	for _, entry := range entries {
		mark, ok := readMark(filepath.Join(marksDir, entry.Name()))
		if ok && mark.Device != thisDevice && now.Sub(mark.Time) < markStaleAfter {
			devices = append(devices, mark.Device)
		}
	}
	return devices
}

// WriteMark leaves device's mark in dir, dated now.
func WriteMark(dir, device string, now time.Time) error {
	path := markPath(dir, device)
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return fmt.Errorf("create device marks folder: %w", err)
	}
	data, err := json.Marshal(deviceMark{Device: device, Time: now})
	if err != nil {
		return fmt.Errorf("encode device mark: %w", err)
	}
	return replaceMark(path, data)
}

// markPath returns where device's mark in dir goes.
func markPath(dir, device string) string {
	return filepath.Join(dir, deviceMarksFolder, device+".json")
}

// KeepMark renews device's mark in dir on each tick, dated by the tick, until
// stop is called. stop ends the renewing and then removes the mark. A mark that
// can't be written or removed is skipped: this is a best guess, so it never
// stops the app, and a mark left behind goes stale.
func KeepMark(dir, device string, ticks <-chan time.Time) (stop func()) {
	stopping := make(chan struct{})
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		for {
			select {
			case now := <-ticks:
				_ = WriteMark(dir, device, now)
			case <-stopping:
				return
			}
		}
	}()
	return func() {
		close(stopping)
		<-stopped
		_ = os.Remove(markPath(dir, device))
	}
}

// replaceMark puts data at path in one step, so that a sync service never
// copies half a mark. It doesn't sync to disk, unlike the store's records: a
// mark lost in a crash costs nothing, as the next one is a minute away.
func replaceMark(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*.tmp")
	if err != nil {
		return fmt.Errorf("create device mark: %w", err)
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write device mark: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close device mark: %w", err)
	}
	if err := os.Rename(tmp.Name(), path); err != nil {
		return fmt.Errorf("replace device mark: %w", err)
	}
	return nil
}

// readMark returns the mark in the file at path, and whether the file is one.
func readMark(path string) (deviceMark, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return deviceMark{}, false
	}
	var mark deviceMark
	if err := json.Unmarshal(data, &mark); err != nil {
		return deviceMark{}, false
	}
	return mark, mark.Device != ""
}
