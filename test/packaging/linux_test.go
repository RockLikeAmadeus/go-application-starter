//go:build packaging && linux

// Package packaging checks the desktop app's installable packages: it builds
// them once with the same task a release uses, then looks inside. It's slow
// and needs the network the first time, so it runs only with the packaging
// build tag (task test:package), not with task check.
package packaging

import (
	"bytes"
	"fmt"
	"image/png"
	"net/mail"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// version is the version the packages are built with, so the tests can check
// it reaches them.
const version = "1.2.3"

// root is the repository's root, where the task runs and the packages land.
const root = "../.."

// appID is the app ID cmd/desktop gives GTK, which names the desktop file and
// the icon.
const appID = "com.example.starter"

// deb, rpm and appImage are the names of the packages the build makes in bin/.
const (
	deb      = "starter-desktop_" + version + "_amd64.deb"
	rpm      = "starter-desktop-" + version + "-1.x86_64.rpm"
	appImage = "starter-desktop-" + version + "-x86_64.AppImage"
)

// linuxErr and windowsErr are why building each platform's packages failed,
// if it did. Every test checks its platform's first, so a failed build fails
// each of its tests with the reason, and only those.
var linuxErr, windowsErr error

func TestMain(m *testing.M) {
	linuxErr = build("package:desktop:linux", version)
	windowsErr = build("package:desktop:windows", version)
	os.Exit(m.Run())
}

// build runs a packaging task with the given VERSION, as a release does. The
// packages land in bin/.
func build(task, version string) error {
	cmd := exec.Command("task", task, "VERSION="+version)
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("task %s VERSION=%s: %w\n%s", task, version, err, out)
	}
	return nil
}

func TestDebInstallsTheProgram(t *testing.T) {
	files := debFiles(t)

	if !slices.Contains(files, "/usr/bin/starter-desktop") {
		t.Errorf("the .deb doesn't install /usr/bin/starter-desktop; it installs %v", files)
	}
}

// Installed files are 0755 (the program) or 0644: only root, who owns them,
// may change them.
func TestDebInstallsFilesOnlyRootCanChange(t *testing.T) {
	out, err := exec.Command("dpkg-deb", "--contents", built(t, deb)).Output()
	if err != nil {
		t.Fatal(err)
	}
	for line := range strings.Lines(string(out)) {
		fields := strings.Fields(line)
		mode, path := fields[0], fields[len(fields)-1]
		if mode[0] != '-' {
			continue // directories
		}
		want := "-rw-r--r--"
		if path == "./usr/bin/starter-desktop" {
			want = "-rwxr-xr-x"
		}
		if mode != want {
			t.Errorf("%s is %s, want %s", strings.TrimPrefix(path, "."), mode, want)
		}
	}
}

// The desktop file puts the app in the app list. Its name is the app ID that
// cmd/desktop gives GTK, which is also the window's Wayland app_id and X11
// WM_CLASS: that's how the desktop matches the window to this file's icon.
func TestDebInstallsTheDesktopFile(t *testing.T) {
	entry := debFile(t, "/usr/share/applications/"+appID+".desktop")

	path := filepath.Join(t.TempDir(), appID+".desktop")
	if err := os.WriteFile(path, entry, 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("desktop-file-validate", path).CombinedOutput(); err != nil {
		t.Errorf("desktop-file-validate: %v\n%s", err, out)
	}
	for _, want := range []string{
		"Name=Starter",
		"Exec=starter-desktop",
		"Icon=" + appID,
		"StartupWMClass=" + appID,
	} {
		if !slices.Contains(strings.Split(string(entry), "\n"), want) {
			t.Errorf("the desktop file has no %q line:\n%s", want, entry)
		}
	}
}

// The desktop file's Icon names the icon by the app ID, and the desktop looks
// it up in the hicolor theme: a 512 × 512 PNG, and the SVG for every other size.
func TestDebInstallsTheIcon(t *testing.T) {
	png512 := debFile(t, "/usr/share/icons/hicolor/512x512/apps/"+appID+".png")
	icon, err := png.Decode(bytes.NewReader(png512))
	if err != nil {
		t.Fatalf("the 512 × 512 icon is not a PNG: %v", err)
	}
	if size := icon.Bounds().Size(); size.X != 512 || size.Y != 512 {
		t.Errorf("the 512 × 512 icon is %v", size)
	}

	svg := debFile(t, "/usr/share/icons/hicolor/scalable/apps/"+appID+".svg")
	want, err := os.ReadFile(filepath.Join(root, "web/public/icon.svg"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(svg, want) {
		t.Error("the scalable icon isn't web/public/icon.svg")
	}
}

// The package is named for the desktop app, so a CLI package could sit
// beside it, carries the version it was built with, and pulls in the GTK 4
// and WebKitGTK 6.0 libraries the program links against.
func TestDebDeclaresItsNameVersionAndDependencies(t *testing.T) {
	if got := debField(t, "Package"); got != "starter-desktop" {
		t.Errorf("Package = %q, want starter-desktop", got)
	}
	if got := debField(t, "Version"); got != version {
		t.Errorf("Version = %q, want %q", got, version)
	}
	depends := strings.Split(debField(t, "Depends"), ", ")
	for _, want := range []string{"libgtk-4-1", "libwebkitgtk-6.0-4"} {
		if !slices.Contains(depends, want) {
			t.Errorf("Depends = %q, want it to include %s", depends, want)
		}
	}
}

// VERSION is `git describe` output unless it's given: a release tag such as
// v2.0.0, or a commit such as 060a096-dirty between releases. Packages need a
// version that starts with a digit, so a tag loses its v and a commit becomes
// 0.0.0.
func TestDebVersionComesFromAReleaseTagOrIsZero(t *testing.T) {
	for given, want := range map[string]string{
		"v2.0.0":        "2.0.0",
		"060a096-dirty": "0.0.0",
	} {
		t.Run(given, func(t *testing.T) {
			// Only the .deb: the AppImage takes minutes to build.
			if err := build("package:desktop:linux:deb", given); err != nil {
				t.Fatal(err)
			}
			pkg := filepath.Join(root, "bin", "starter-desktop_"+want+"_amd64.deb")
			if got := controlField(t, pkg, "Version"); got != want {
				t.Errorf("Version = %q, want %q", got, want)
			}
		})
	}
}

// Debian requires a maintainer, as "Name <email>". Without one, nfpm warns
// and writes a placeholder, "Unset Maintainer <unset@localhost>".
func TestDebNamesAMaintainer(t *testing.T) {
	got := debField(t, "Maintainer")
	addr, err := mail.ParseAddress(got)
	if err != nil || addr.Name == "" || strings.HasSuffix(addr.Address, "@localhost") {
		t.Errorf("Maintainer = %q, want a name and a real email address", got)
	}
}

// After installing and after removing, the scripts refresh the app list and
// the icon cache, so the app appears and disappears at once. Systems without
// those tools must still install and remove it.
func TestDebScriptsRefreshTheAppListAndIcons(t *testing.T) {
	for _, script := range []string{"postinst", "postrm"} {
		t.Run(script, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), script)
			if err := os.WriteFile(path, debControlFile(t, script), 0o755); err != nil {
				t.Fatal(err)
			}

			// Stub tools that log their arguments.
			bin, log := t.TempDir(), filepath.Join(t.TempDir(), "calls")
			for _, tool := range []string{"update-desktop-database", "gtk-update-icon-cache"} {
				stub := "#!/bin/sh\necho " + tool + " \"$@\" >> " + log + "\n"
				if err := os.WriteFile(filepath.Join(bin, tool), []byte(stub), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			run := exec.Command("/bin/sh", path, "configure")
			run.Env = []string{"PATH=" + bin}
			if out, err := run.CombinedOutput(); err != nil {
				t.Fatalf("%s with the tools: %v\n%s", script, err, out)
			}
			calls, _ := os.ReadFile(log)
			for _, want := range []string{
				"update-desktop-database -q /usr/share/applications",
				"gtk-update-icon-cache -q -t -f /usr/share/icons/hicolor",
			} {
				if !slices.Contains(strings.Split(string(calls), "\n"), want) {
					t.Errorf("%s didn't run %q; it ran:\n%s", script, want, calls)
				}
			}

			without := exec.Command("/bin/sh", path, "configure")
			without.Env = []string{"PATH=" + t.TempDir()}
			if out, err := without.CombinedOutput(); err != nil {
				t.Errorf("%s without the tools: %v\n%s", script, err, out)
			}
		})
	}
}

// The .rpm is built from the same nfpm.yaml, so Fedora and openSUSE get the
// same app as Debian and Ubuntu.
func TestRpmInstallsTheSameFilesAsTheDeb(t *testing.T) {
	pkg := builtRpm(t)
	out, err := exec.Command("rpm", "--query", "--list", "--package", pkg).Output()
	if err != nil {
		t.Fatalf("rpm --query --list %s: %v", pkg, err)
	}
	got := strings.Fields(string(out))
	want := debFiles(t)
	slices.Sort(got)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("the .rpm installs %v, want the .deb's %v", got, want)
	}
}

// Fedora and openSUSE name the GTK 4 and WebKitGTK 6.0 packages differently
// from Debian, so the .rpm must require theirs.
func TestRpmRequiresTheRpmNamesOfItsLibraries(t *testing.T) {
	pkg := builtRpm(t)
	out, err := exec.Command("rpm", "--query", "--requires", "--package", pkg).Output()
	if err != nil {
		t.Fatalf("rpm --query --requires %s: %v", pkg, err)
	}
	requires := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, want := range []string{"gtk4", "webkitgtk6.0"} {
		if !slices.Contains(requires, want) {
			t.Errorf("the .rpm requires %q, want it to include %s", requires, want)
		}
	}
	for _, debOnly := range []string{"libgtk-4-1", "libwebkitgtk-6.0-4"} {
		if slices.Contains(requires, debOnly) {
			t.Errorf("the .rpm requires the Debian package %s", debOnly)
		}
	}
}

// The AppImage runs without installing, on distributions with neither dpkg
// nor rpm. It carries the program, its desktop file and icon at the top
// (where AppImage tools look for them), and the libraries it needs.
func TestAppImageHoldsTheProgramDesktopFileAndIcon(t *testing.T) {
	image, err := filepath.Abs(built(t, appImage))
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	extract := exec.Command(image, "--appimage-extract")
	extract.Dir = dir
	if out, err := extract.CombinedOutput(); err != nil {
		t.Fatalf("%s --appimage-extract: %v\n%s", appImage, err, out)
	}
	for _, file := range []string{
		"usr/bin/starter-desktop",
		appID + ".desktop",
		appID + ".png",
	} {
		if _, err := os.Stat(filepath.Join(dir, "squashfs-root", file)); err != nil {
			t.Errorf("the AppImage has no %s: %v", file, err)
		}
	}
}

// debControlFile returns one file of the .deb's control archive, such as its
// postinst script.
func debControlFile(t *testing.T, name string) []byte {
	t.Helper()
	pkg := built(t, deb)
	out, err := exec.Command("sh", "-c", `dpkg-deb --ctrl-tarfile "$1" | tar -xO "./$2"`, "sh", pkg, name).Output()
	if err != nil {
		t.Fatalf("the .deb has no %s script: %v", name, err)
	}
	return out
}

// debField returns one field of the .deb's control file, or "" if it has none.
func debField(t *testing.T, name string) string {
	t.Helper()
	return controlField(t, built(t, deb), name)
}

// controlField returns one field of a .deb's control file, or "" if it has
// none.
func controlField(t *testing.T, pkg, name string) string {
	t.Helper()
	out, err := exec.Command("dpkg-deb", "--field", pkg, name).Output()
	if err != nil {
		t.Fatalf("dpkg-deb --field %s %s: %v", pkg, name, err)
	}
	return strings.TrimSpace(string(out))
}

// debFile returns the contents of one file the .deb installs.
func debFile(t *testing.T, file string) []byte {
	t.Helper()
	pkg := built(t, deb)
	out, err := exec.Command("sh", "-c", `dpkg-deb --fsys-tarfile "$1" | tar -xO ".$2"`, "sh", pkg, file).Output()
	if err != nil {
		t.Fatalf("the .deb has no %s: %v", file, err)
	}
	return out
}

// debFiles lists the files the .deb installs, as absolute paths.
func debFiles(t *testing.T) []string {
	t.Helper()
	pkg := built(t, deb)
	out, err := exec.Command("dpkg-deb", "--contents", pkg).Output()
	if err != nil {
		t.Fatalf("dpkg-deb --contents %s: %v", pkg, err)
	}
	var files []string
	for line := range strings.Lines(string(out)) {
		// Each line ends with the path, as ./usr/bin/…; directories end in /.
		fields := strings.Fields(line)
		path := strings.TrimPrefix(fields[len(fields)-1], ".")
		if !strings.HasSuffix(path, "/") {
			files = append(files, path)
		}
	}
	return files
}

// builtRpm returns the .rpm's path, skipping the test where the rpm tool,
// which reads it, isn't installed.
func builtRpm(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("rpm"); err != nil {
		t.Skip("rpm isn't installed (sudo apt install rpm)")
	}
	return built(t, rpm)
}

// built returns the path of a Linux package the build made in bin/, failing
// the test if the build failed or didn't make it.
func built(t *testing.T, name string) string {
	t.Helper()
	return builtBy(t, linuxErr, name)
}

// builtBy returns the path of a file a build made in bin/, failing the test
// if that build failed (buildErr) or didn't make it.
func builtBy(t *testing.T, buildErr error, name string) string {
	t.Helper()
	if buildErr != nil {
		t.Fatal(buildErr)
	}
	path := filepath.Join(root, "bin", name)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the build didn't make bin/%s: %v", name, err)
	}
	return path
}
