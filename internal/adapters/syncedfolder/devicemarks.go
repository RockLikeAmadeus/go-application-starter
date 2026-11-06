package syncedfolder

import (
	"encoding/json"
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
// skewed clock can miss a device or keep a stale one.
func OtherDevices(dir, thisDevice string, now time.Time) []string {
	marksDir := filepath.Join(dir, deviceMarksFolder)
	entries, _ := os.ReadDir(marksDir)
	var devices []string
	for _, entry := range entries {
		var mark deviceMark
		data, _ := os.ReadFile(filepath.Join(marksDir, entry.Name()))
		_ = json.Unmarshal(data, &mark)
		if mark.Device != thisDevice && now.Sub(mark.Time) < markStaleAfter {
			devices = append(devices, mark.Device)
		}
	}
	return devices
}
