package syncedfolder

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// deviceMarksFolder holds a file for each device that marked the folder.
const deviceMarksFolder = ".devices"

// A deviceMark is the file a device leaves to say it has the folder in use.
type deviceMark struct {
	Device string `json:"device"`
}

// OtherDevices returns the devices other than thisDevice that have marked dir
// as in use.
func OtherDevices(dir, thisDevice string, now time.Time) []string {
	marksDir := filepath.Join(dir, deviceMarksFolder)
	entries, _ := os.ReadDir(marksDir)
	var devices []string
	for _, entry := range entries {
		var mark deviceMark
		data, _ := os.ReadFile(filepath.Join(marksDir, entry.Name()))
		_ = json.Unmarshal(data, &mark)
		if mark.Device != thisDevice {
			devices = append(devices, mark.Device)
		}
	}
	return devices
}
