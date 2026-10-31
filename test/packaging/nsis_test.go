//go:build packaging && linux

package packaging

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
	"unicode/utf16"
)

// exe is the desktop app's Windows program, cross-compiled into bin/, and
// setup is the setup program that installs it.
const (
	exe   = "starter-desktop.exe"
	setup = "starter-desktop-" + version + "-setup.exe"
)

// The program is a GUI program for 64-bit Windows, so it opens no console
// window beside its own.
func TestExeIsAWindowsGUIProgramForAmd64(t *testing.T) {
	f := openWindows(t, exe)

	if f.Machine != pe.IMAGE_FILE_MACHINE_AMD64 {
		t.Errorf("bin/%s is for machine %#x, want amd64 (%#x)", exe, f.Machine, pe.IMAGE_FILE_MACHINE_AMD64)
	}
	header, ok := f.OptionalHeader.(*pe.OptionalHeader64)
	if !ok {
		t.Fatalf("bin/%s has a %T, want a 64-bit optional header", exe, f.OptionalHeader)
	}
	if header.Subsystem != pe.IMAGE_SUBSYSTEM_WINDOWS_GUI {
		t.Errorf("bin/%s is for subsystem %d, want the Windows GUI (%d)", exe, header.Subsystem, pe.IMAGE_SUBSYSTEM_WINDOWS_GUI)
	}
}

// The program carries the product's icon, so Explorer, the Start menu and the
// taskbar show it: its icon images are the ones in build/desktop/windows/icon.ico.
func TestExeCarriesTheProductIcon(t *testing.T) {
	f := openWindows(t, exe)

	got := resources(t, f)[rtIcon]
	if len(got) == 0 {
		t.Fatalf("bin/%s has no icon", exe)
	}
	want := icoImages(t, filepath.Join(root, "build/desktop/windows/icon.ico"))
	slices.SortFunc(got, bytes.Compare)
	slices.SortFunc(want, bytes.Compare)
	if !slices.EqualFunc(got, want, bytes.Equal) {
		t.Errorf("bin/%s has %d icon images that aren't icon.ico's %d", exe, len(got), len(want))
	}
}

// The program's and the setup program's version info name the product and
// the package version, as Explorer's Properties → Details shows them. Windows
// versions have four numbers, so 1.2.3 is 1.2.3.0.
func TestVersionInfoNamesTheProductAndVersion(t *testing.T) {
	for _, name := range []string{exe, setup} {
		t.Run(name, func(t *testing.T) {
			fileVersion, productVersion, text := windowsVersionInfo(t, name)

			if want := version + ".0"; fileVersion != want || productVersion != want {
				t.Errorf("bin/%s's file and product versions are %s and %s, want %s", name, fileVersion, productVersion, want)
			}
			if got := text["ProductName"]; got != "Starter" {
				t.Errorf("bin/%s's product name is %q, want Starter", name, got)
			}
			if got := text["ProductVersion"]; got != version {
				t.Errorf("bin/%s's product version string is %q, want %s", name, got, version)
			}
		})
	}
}

// The program's version info describes it and names a publisher. Task
// Manager shows the description as the process's name, and Properties →
// Details shows all three. The publisher and copyright are placeholders that
// a project fills in.
func TestExeVersionInfoDescribesTheProgramAndItsPublisher(t *testing.T) {
	_, _, text := windowsVersionInfo(t, exe)

	if got := text["FileDescription"]; got != "Starter" {
		t.Errorf("bin/%s's file description is %q, want Starter", exe, got)
	}
	for _, key := range []string{"CompanyName", "LegalCopyright"} {
		if text[key] == "" {
			t.Errorf("bin/%s's version info has no %s", exe, key)
		}
	}
}

// The setup program names the same publisher and copyright as the program
// it installs, so a project fills them in once.
func TestSetupNamesTheSamePublisherAsTheProgram(t *testing.T) {
	_, _, program := windowsVersionInfo(t, exe)
	_, _, installer := windowsVersionInfo(t, setup)

	for _, key := range []string{"CompanyName", "LegalCopyright"} {
		if installer[key] != program[key] {
			t.Errorf("bin/%s's %s is %q, but bin/%s's is %q", setup, key, installer[key], exe, program[key])
		}
	}
}

// The setup program carries the program, byte for byte the one built, and
// Microsoft's WebView2 bootstrapper, which installs the runtime the program
// needs where Windows lacks it.
func TestSetupHoldsTheProgramAndTheWebView2Bootstrapper(t *testing.T) {
	sevenZip := sevenZip(t)
	path := builtForWindows(t, setup)

	program, err := exec.Command(sevenZip, "e", "-so", path, exe).Output()
	if err != nil {
		t.Fatalf("7-Zip can't take %s out of bin/%s: %v", exe, setup, err)
	}
	if built, err := os.ReadFile(builtForWindows(t, exe)); err != nil {
		t.Fatal(err)
	} else if !bytes.Equal(program, built) {
		t.Errorf("bin/%s holds a %s that isn't the one built (%d bytes, want %d)", setup, exe, len(program), len(built))
	}

	listing, err := exec.Command(sevenZip, "l", "-slt", path).Output()
	if err != nil {
		t.Fatal(err)
	}
	const bootstrapper = "Path = $PLUGINSDIR/webview2bootstrapper/MicrosoftEdgeWebview2Setup.exe"
	if !bytes.Contains(listing, []byte(bootstrapper)) {
		t.Errorf("bin/%s doesn't hold the WebView2 bootstrapper; 7-Zip lists:\n%s", setup, listing)
	}
}

// The setup program installs for the current user only: Windows asks for no
// administrator (its manifest asks to run as the user who started it), and it
// installs into that user's own programs folder.
func TestSetupInstallsForTheCurrentUser(t *testing.T) {
	manifests := resources(t, openWindows(t, setup))[rtManifest]
	if len(manifests) != 1 {
		t.Fatalf("bin/%s has %d manifests, want 1", setup, len(manifests))
	}
	level := "no execution level"
	if found := executionLevel.FindSubmatch(manifests[0]); found != nil {
		level = string(found[1])
	}
	if level != "asInvoker" {
		t.Errorf("bin/%s asks Windows to run it with %s, want asInvoker (no administrator)", setup, level)
	}

	if dir := readMakensisLog(t).installDir(); !strings.HasPrefix(dir, `$LOCALAPPDATA\Programs\`) {
		t.Errorf("bin/%s installs into %s, want a folder under $LOCALAPPDATA\\Programs", setup, dir)
	}
}

// The setup program puts the app in the Start menu under the product's name,
// and leaves the desktop alone: a user who wants a desktop shortcut makes one.
func TestSetupMakesAStartMenuShortcutOnly(t *testing.T) {
	shortcuts := readMakensisLog(t).shortcuts()

	want := []string{`$SMPROGRAMS\Starter.lnk`}
	if !slices.Equal(shortcuts, want) {
		t.Errorf("bin/%s makes the shortcuts %q, want %q", setup, shortcuts, want)
	}
}

// The uninstaller removes what the setup program made, including WebView2's
// own folder, but leaves the user's items in their data folder (SPEC.md §4.6),
// and touches nothing on the desktop, where the setup program puts nothing.
func TestUninstallerKeepsTheUsersData(t *testing.T) {
	removed := readMakensisLog(t).uninstallerRemoves()

	// WebView2 keeps its cache and the web app's local storage (the chosen
	// style and theme) in %AppData%\<program name>.
	if webView2 := `$AppData\starter-desktop.exe`; !slices.Contains(removed, webView2) {
		t.Errorf("the uninstaller doesn't remove WebView2's folder, %s; it removes %q", webView2, removed)
	}
	for _, path := range removed {
		// The items live in %AppData%\starter (os.UserConfigDir()\starter).
		if strings.EqualFold(path, `$AppData\starter`) || strings.HasPrefix(strings.ToLower(path), strings.ToLower(`$AppData\starter\`)) {
			t.Errorf("the uninstaller removes %s, where the user's items are", path)
		}
		if strings.HasPrefix(path, `$DESKTOP\`) {
			t.Errorf("the uninstaller removes %s from the desktop, where the setup program puts nothing", path)
		}
	}
}

// executionLevel finds the level a manifest asks Windows to run a program at.
var executionLevel = regexp.MustCompile(`requestedExecutionLevel\s+level="([^"]+)"`)

// makensisLog is what makensis printed (at -V4, one line per command) when
// it built the setup program. The build keeps it beside the program, because
// the setup program's own copy of its script is compressed.
type makensisLog string

// makensisLogFile is where the build keeps makensis's log, under bin/.
const makensisLogFile = "windows/makensis.log"

func readMakensisLog(t *testing.T) makensisLog {
	t.Helper()
	data, err := os.ReadFile(builtForWindows(t, makensisLogFile))
	if err != nil {
		t.Fatal(err)
	}
	return makensisLog(data)
}

// installDir is the setup program's default install folder: makensis logs
// it as InstallDir: "<folder>".
func (l makensisLog) installDir() string {
	for line := range strings.Lines(string(l)) {
		if dir, ok := strings.CutPrefix(strings.TrimSpace(line), "InstallDir: "); ok {
			return strings.Trim(dir, `"`)
		}
	}
	return ""
}

// shortcuts are the shortcuts the setup program makes: makensis logs each as
// CreateShortcut: "<shortcut>"->"<target>" ….
func (l makensisLog) shortcuts() []string {
	var found []string
	for line := range strings.Lines(string(l)) {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), `CreateShortcut: "`); ok {
			shortcut, _, _ := strings.Cut(rest, `"`)
			found = append(found, shortcut)
		}
	}
	return found
}

// uninstallerRemoves are the files and folders the uninstaller deletes:
// makensis logs its section as Section: "uninstall" … SectionEnd, with each
// RMDir: [/r] "<folder>" and Delete: "<file>" inside.
func (l makensisLog) uninstallerRemoves() []string {
	var found []string
	inUninstall := false
	for line := range strings.Lines(string(l)) {
		line = strings.TrimSpace(line)
		switch {
		case line == `Section: "uninstall"`:
			inUninstall = true
		case line == "SectionEnd":
			inUninstall = false
		case inUninstall:
			for _, prefix := range []string{`RMDir: /r "`, `RMDir: "`, `Delete: "`} {
				if rest, ok := strings.CutPrefix(line, prefix); ok {
					path, _, _ := strings.Cut(rest, `"`)
					found = append(found, path)
				}
			}
		}
	}
	return found
}

// sevenZip returns the 7-Zip command, which reads NSIS installers, skipping
// the test where it isn't installed.
func sevenZip(t *testing.T) string {
	t.Helper()
	for _, name := range []string{"7z", "7zz"} {
		if path, err := exec.LookPath(name); err == nil {
			return path
		}
	}
	t.Skip("7-Zip isn't installed (sudo apt install 7zip)")
	return ""
}

// rtIcon, rtVersion and rtManifest are the resource types of one icon image
// (RT_ICON), the version info (RT_VERSION) and the manifest (RT_MANIFEST).
const (
	rtIcon     = 3
	rtVersion  = 16
	rtManifest = 24
)

// resources reads a Windows program's resources (its .rsrc section) by type,
// gathering every name and language of each type.
func resources(t *testing.T, f *pe.File) map[uint32][][]byte {
	t.Helper()
	section := f.Section(".rsrc")
	if section == nil {
		return nil
	}
	data, err := section.Data()
	if err != nil {
		t.Fatal(err)
	}
	// The resources are a three-level tree: type, then name, then language.
	// Each directory is a 16-byte header, then 8-byte entries, named ones
	// first; an entry whose offset has the high bit set is a subdirectory.
	type entry struct{ id, offset uint32 }
	entries := func(offset uint32) []entry {
		named := binary.LittleEndian.Uint16(data[offset+12:])
		ids := binary.LittleEndian.Uint16(data[offset+14:])
		var list []entry
		for i := range uint32(named) + uint32(ids) {
			at := offset + 16 + 8*i
			list = append(list, entry{binary.LittleEndian.Uint32(data[at:]), binary.LittleEndian.Uint32(data[at+4:])})
		}
		return list
	}
	const subdirectory = 1 << 31
	found := map[uint32][][]byte{}
	for _, typ := range entries(0) {
		for _, name := range entries(typ.offset &^ subdirectory) {
			for _, lang := range entries(name.offset &^ subdirectory) {
				// A leaf: the data's address in memory (RVA) and its size.
				rva := binary.LittleEndian.Uint32(data[lang.offset:])
				size := binary.LittleEndian.Uint32(data[lang.offset+4:])
				start := rva - section.VirtualAddress
				found[typ.id] = append(found[typ.id], data[start:start+size])
			}
		}
	}
	return found
}

// windowsVersionInfo reads the one version info of a Windows program the
// build made in bin/ (see versionInfo).
func windowsVersionInfo(t *testing.T, name string) (fileVersion, productVersion string, text map[string]string) {
	t.Helper()
	infos := resources(t, openWindows(t, name))[rtVersion]
	if len(infos) != 1 {
		t.Fatalf("bin/%s has %d version infos, want 1", name, len(infos))
	}
	return versionInfo(t, infos[0])
}

// versionInfo reads a VS_VERSIONINFO resource: its fixed file and product
// versions (a.b.c.d) and its strings (ProductName and so on).
//
// It's a tree of blocks. Each is its length, its value's length, its type (1
// for text), a NUL-terminated UTF-16 key, padding to 4 bytes, the value,
// padding to 4 bytes, then its child blocks. The root's value is the
// VS_FIXEDFILEINFO; the strings are leaves under StringFileInfo.
func versionInfo(t *testing.T, data []byte) (fileVersion, productVersion string, text map[string]string) {
	t.Helper()
	u16 := func(at int) int { return int(binary.LittleEndian.Uint16(data[at:])) }
	u32 := func(at int) uint32 { return binary.LittleEndian.Uint32(data[at:]) }
	align := func(at int) int { return (at + 3) &^ 3 }
	// readText reads UTF-16 from at up to a NUL or end, returning it and
	// where the NUL ends.
	readText := func(at, end int) (string, int) {
		var units []uint16
		for ; at+1 < end; at += 2 {
			unit := uint16(u16(at))
			if unit == 0 {
				return string(utf16.Decode(units)), at + 2
			}
			units = append(units, unit)
		}
		return string(utf16.Decode(units)), end
	}
	versionAt := func(at int) string {
		ms, ls := u32(at), u32(at+4)
		return fmt.Sprintf("%d.%d.%d.%d", ms>>16, ms&0xffff, ls>>16, ls&0xffff)
	}

	text = map[string]string{}
	var walk func(at int)
	walk = func(at int) {
		end := at + u16(at)
		valueLength, isText := u16(at+2), u16(at+4) == 1
		key, valueAt := readText(at+6, end)
		valueAt = align(valueAt)
		switch {
		case key == "VS_VERSION_INFO" && valueLength > 0:
			// VS_FIXEDFILEINFO: signature, struct version, then the file and
			// product versions, each as two 32-bit halves.
			if u32(valueAt) != 0xfeef04bd {
				t.Fatalf("version info has no VS_FIXEDFILEINFO")
			}
			fileVersion, productVersion = versionAt(valueAt+8), versionAt(valueAt+16)
		case isText && valueLength > 0:
			text[key], _ = readText(valueAt, end)
			return
		}
		valueSize := valueLength
		if isText {
			valueSize *= 2 // a text value's length counts UTF-16 units
		}
		for child := align(valueAt + valueSize); child < end && u16(child) > 0; child = align(child + u16(child)) {
			walk(child)
		}
	}
	walk(0)
	return fileVersion, productVersion, text
}

// icoImages reads the images in a .ico file: a 6-byte header with the count,
// then a 16-byte entry per image holding its size and offset.
func icoImages(t *testing.T, path string) [][]byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var images [][]byte
	for i := range int(binary.LittleEndian.Uint16(data[4:])) {
		at := 6 + 16*i
		size := binary.LittleEndian.Uint32(data[at+8:])
		offset := binary.LittleEndian.Uint32(data[at+12:])
		images = append(images, data[offset:offset+size])
	}
	return images
}

// openWindows opens a Windows program the build made in bin/, closing it
// when the test ends.
func openWindows(t *testing.T, name string) *pe.File {
	t.Helper()
	f, err := pe.Open(builtForWindows(t, name))
	if err != nil {
		t.Fatalf("bin/%s isn't a Windows program: %v", name, err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

// builtForWindows returns the path of a file the Windows build made in bin/,
// failing the test if the build failed or didn't make it.
func builtForWindows(t *testing.T, name string) string {
	t.Helper()
	return builtBy(t, windowsErr, name)
}
