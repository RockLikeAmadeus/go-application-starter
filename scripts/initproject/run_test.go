package main_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	initproject "github.com/RockLikeAmadeus/go-application-starter/scripts/initproject"
	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

func TestRunRefusesBeforeTouchingFiles(t *testing.T) {
	const original = "module github.com/RockLikeAmadeus/go-application-starter\n"
	valid := withDetails("--module", "example.com/acme/tasks", "--name", "Acme")
	tests := []struct {
		desc      string
		args      []string
		porcelain string
		wantErr   string
	}{
		{desc: "missing module", args: withDetails("--name", "Acme"), wantErr: "--module is required"},
		{desc: "invalid module", args: withDetails("--module", "not a module", "--name", "Acme"), wantErr: "invalid module path"},
		{desc: "unusable name", args: withDetails("--module", "example.com/acme/tasks", "--name", "!!!"), wantErr: "name must contain a letter or digit"},
		{desc: "uncommitted changes", args: valid, porcelain: " M go.mod\n", wantErr: "commit or stash your changes first"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			dir := t.TempDir()
			gomod := filepath.Join(dir, "go.mod")
			if err := os.WriteFile(gomod, []byte(original), 0o644); err != nil {
				t.Fatal(err)
			}

			_, err := initproject.Run(dir, nil, tt.args, tt.porcelain)

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Run error = %v, want one mentioning %q", err, tt.wantErr)
			}
			if got, _ := os.ReadFile(gomod); string(got) != original {
				t.Errorf("go.mod changed to %q", got)
			}
		})
	}
}

func TestRunRenamesTheProject(t *testing.T) {
	dir := t.TempDir()
	writeTemplateRecord(t, dir)
	write := func(rel, content string) { writeFile(t, dir, rel, content) }
	read := func(rel string) string { return readFile(t, dir, rel) }
	const goSum = "github.com/RockLikeAmadeus/go-application-starter v0.0.0 h1:x=\n"
	const ignored = "// Starter\n"
	write("go.mod", "module github.com/RockLikeAmadeus/go-application-starter\n")
	write("cmd/starter/main.go", "// Command starter is the Starter CLI.\n")
	write("go.sum", goSum)
	write("web/node_modules/x.js", ignored) // not tracked, so not listed
	files := []string{"go.mod", "cmd/starter/main.go", "go.sum"}

	_, err := initproject.Run(dir, files, withDetails("--module", "example.com/acme/tasks", "--name", "Acme Tasks"), "")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	if got, want := read("go.mod"), "module example.com/acme/tasks\n"; got != want {
		t.Errorf("go.mod = %q, want %q", got, want)
	}
	if got, want := read("cmd/acme-tasks/main.go"), "// Command acme-tasks is the Acme Tasks CLI.\n"; got != want {
		t.Errorf("cmd/acme-tasks/main.go = %q, want %q", got, want)
	}
	if _, err := os.Stat(filepath.Join(dir, "cmd", "starter")); !os.IsNotExist(err) {
		t.Errorf("cmd/starter still exists (stat error %v)", err)
	}
	if got := read("go.sum"); got != goSum {
		t.Errorf("go.sum was rewritten to %q", got)
	}
	if got := read("web/node_modules/x.js"); got != ignored {
		t.Errorf("an unlisted file was rewritten to %q", got)
	}
}

func TestRunRemovesTheSampleOnlyWhenAsked(t *testing.T) {
	const bootstrap = "package bootstrap\n\n" +
		"// sample:start\n" +
		"var items = 1\n\n" +
		"// sample:end\n" +
		"var health = 2\n"
	tests := []struct {
		desc          string
		flags         []string
		wantSample    bool
		wantBootstrap string
	}{
		{desc: "kept by default", wantSample: true, wantBootstrap: bootstrap},
		{
			desc:          "removed with --remove-sample",
			flags:         []string{"--remove-sample"},
			wantSample:    false,
			wantBootstrap: "package bootstrap\n\nvar health = 2\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			dir := t.TempDir()
			writeTemplateRecord(t, dir)
			files := map[string]string{
				"internal/core/domain/item.go":         "package domain\n",
				"web/src/features/items/ItemsPage.tsx": "export {};\n",
				"internal/bootstrap/bootstrap.go":      bootstrap,
			}
			var listed []string
			for rel, content := range files {
				writeFile(t, dir, rel, content)
				listed = append(listed, rel)
			}
			args := append(withDetails("--module", "example.com/acme/tasks", "--name", "Acme"), tt.flags...)

			if _, err := initproject.Run(dir, listed, args, ""); err != nil {
				t.Fatalf("Run error = %v", err)
			}

			for _, rel := range []string{"internal/core/domain/item.go", "web/src/features/items"} {
				_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel)))
				if isPresent := err == nil; isPresent != tt.wantSample {
					t.Errorf("%s exists = %v, want %v", rel, isPresent, tt.wantSample)
				}
			}
			if got := readFile(t, dir, "internal/bootstrap/bootstrap.go"); got != tt.wantBootstrap {
				t.Errorf("bootstrap.go =\n%q\nwant\n%q", got, tt.wantBootstrap)
			}
		})
	}
}

func TestRunRemovesTheInitTool(t *testing.T) {
	dir := t.TempDir()
	writeTemplateRecord(t, dir)
	files := map[string]string{
		"README.md": "# Starter\n" +
			"<!-- template:start -->\n" +
			"## Using this template\n" +
			"<!-- template:end -->\n" +
			"## Run it\n",
		"Taskfile.yml": "tasks:\n" +
			"  # template:start\n" +
			"  init:\n" +
			"    cmds: [go run ./scripts/initproject]\n" +
			"  # template:end\n" +
			"  test:\n",
		"docs/template.md": "# Using this template\n",
		// The tool's own tests mention markers without closing them; they must
		// be deleted, not processed.
		"scripts/initproject/blocks_test.go": `in := "// template:start\n"` + "\n",
	}
	var listed []string
	for rel, content := range files {
		writeFile(t, dir, rel, content)
		listed = append(listed, rel)
	}

	_, err := initproject.Run(dir, listed, withDetails("--module", "example.com/acme/tasks", "--name", "Acme"), "")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	for _, rel := range []string{"scripts/initproject", "docs/template.md"} {
		if _, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel))); !os.IsNotExist(err) {
			t.Errorf("%s still exists (stat error %v)", rel, err)
		}
	}
	if got, want := readFile(t, dir, "README.md"), "# Acme\n## Run it\n"; got != want {
		t.Errorf("README.md = %q, want %q", got, want)
	}
	if got, want := readFile(t, dir, "Taskfile.yml"), "tasks:\n  test:\n"; got != want {
		t.Errorf("Taskfile.yml = %q, want %q", got, want)
	}
}

func TestRunTidiesGoFilesAfterRemovingTheSample(t *testing.T) {
	dir := t.TempDir()
	writeTemplateRecord(t, dir)
	writeFile(t, dir, "internal/core/ports/ports.go", `package ports

import (
	"context"
	"fmt"
	"time"
)

// sample:start
type ItemService interface {
	ListItems(ctx context.Context) error
}

// sample:end

// Clock abstracts time.
type Clock interface {
	Now() time.Time
}

var _ = fmt.Sprint
`)

	_, err := initproject.Run(dir, []string{"internal/core/ports/ports.go"},
		withDetails("--module", "example.com/acme/tasks", "--name", "Acme", "--remove-sample"), "")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	want := `package ports

import (
	"fmt"
	"time"
)

// Clock abstracts time.
type Clock interface {
	Now() time.Time
}

var _ = fmt.Sprint
`
	if got := readFile(t, dir, "internal/core/ports/ports.go"); got != want {
		t.Errorf("ports.go =\n%s\nwant\n%s", got, want)
	}
}

func TestRunAsksForFollowUpCommands(t *testing.T) {
	tests := []struct {
		desc  string
		flags []string
		want  [][]string
	}{
		{
			desc: "sample kept: go.mod drops the init tool's own dependencies",
			want: [][]string{{"go", "mod", "tidy"}},
		},
		{
			desc:  "sample removed: API code and go.mod follow the trimmed sources",
			flags: []string{"--remove-sample"},
			want: [][]string{
				{"go", "tool", "-modfile=tools/go.mod", "oapi-codegen", "-config", "api/oapi-codegen.server.yaml", "api/openapi.yaml"},
				{"go", "tool", "-modfile=tools/go.mod", "oapi-codegen", "-config", "api/oapi-codegen.client.yaml", "api/openapi.yaml"},
				{"go", "mod", "tidy"},
				{"pnpm", "--dir", "web", "gen:api"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			args := append(withDetails("--module", "example.com/acme/tasks", "--name", "Acme"), tt.flags...)

			dir := t.TempDir()
			writeTemplateRecord(t, dir)

			got, err := initproject.Run(dir, nil, args, "")
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}

			if !slices.EqualFunc(got, tt.want, slices.Equal) {
				t.Errorf("follow-up commands =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

func TestRunRemovesDirectoriesTheSampleLeavesEmpty(t *testing.T) {
	dir := t.TempDir()
	writeTemplateRecord(t, dir)
	files := []string{
		"internal/core/app/item_service.go", // the directory's only file
		"internal/core/domain/item.go",
		"internal/core/domain/errors.go", // not sample, so domain/ stays
	}
	for _, rel := range files {
		writeFile(t, dir, rel, "package x\n")
	}
	args := withDetails("--module", "example.com/acme/tasks", "--name", "Acme", "--remove-sample")

	if _, err := initproject.Run(dir, files, args, ""); err != nil {
		t.Fatalf("Run error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, "internal", "core", "app")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("internal/core/app still exists (stat error %v), want it removed once empty", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "internal", "core", "domain", "errors.go")); err != nil {
		t.Errorf("internal/core/domain/errors.go: %v, want it kept", err)
	}
}

func TestRunKeepsGoImportsSortedWhenTheModuleSortsDifferently(t *testing.T) {
	dir := t.TempDir()
	writeTemplateRecord(t, dir)
	// The template's module sorts before github.com/spf13; zenith.dev sorts after.
	writeFile(t, dir, "cmd/starter/main.go", "package main\n\nimport (\n"+
		"\t\"github.com/RockLikeAmadeus/go-application-starter/internal/cli\"\n"+
		"\t\"github.com/spf13/cobra\"\n"+
		")\n\nvar _, _ = cli.New, cobra.Command{}\n")

	_, err := initproject.Run(dir, []string{"cmd/starter/main.go"},
		withDetails("--module", "zenith.dev/notes", "--name", "Zenith"), "")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	want := "package main\n\nimport (\n" +
		"\t\"github.com/spf13/cobra\"\n" +
		"\t\"zenith.dev/notes/internal/cli\"\n" +
		")\n\nvar _, _ = cli.New, cobra.Command{}\n"
	if got := readFile(t, dir, "cmd/zenith/main.go"); got != want {
		t.Errorf("cmd/zenith/main.go =\n%s\nwant\n%s", got, want)
	}
}

func TestRunReadsTheOldNamesFromTheProjectRecord(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "project.json",
		`{"module": "example.com/old/thing", "name": "Old Thing", "slug": "old-thing", "app": "old-thing"}`+"\n")
	writePackagingFiles(t, dir)
	writeFile(t, dir, "README.md", "# Old Thing\n\nOLD_THING_PORT=8080 go run example.com/old/thing/cmd/old-thing\n")

	_, err := initproject.Run(dir, []string{"README.md"},
		withDetails("--module", "example.com/acme/tasks", "--name", "Acme"), "")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	want := "# Acme\n\nACME_PORT=8080 go run example.com/acme/tasks/cmd/acme\n"
	if got := readFile(t, dir, "README.md"); got != want {
		t.Errorf("README.md = %q, want %q", got, want)
	}
}

func TestRunRecordsTheNewNames(t *testing.T) {
	dir := t.TempDir()
	writePackagingFiles(t, dir)
	// The old names are also keys in project.json, so renaming its text
	// alone would rename the keys too.
	writeFile(t, dir, "project.json",
		`{"module": "example.com/you/app", "name": "App", "slug": "app", "app": "app"}`+"\n")

	_, err := initproject.Run(dir, []string{"project.json"},
		withDetails("--module", "example.com/acme/tasks", "--name", "Acme Tasks"), "")
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	got, err := project.ReadRecord(dir)
	if err != nil {
		t.Fatalf("project.json is unreadable after Run: %v\n%s", err, readFile(t, dir, "project.json"))
	}
	want := project.Names{Module: "example.com/acme/tasks", Brand: "Acme Tasks", Slug: "acme-tasks", App: "acme-tasks", EnvPrefix: "ACME_TASKS_"}
	if got != want {
		t.Errorf("project.json records %+v, want %+v\n%s", got, want, readFile(t, dir, "project.json"))
	}
}

func TestRunSetsThePackageDetails(t *testing.T) {
	const (
		nfpm    = "build/desktop/linux/nfpm.yaml"
		desktop = "build/desktop/linux/app.desktop"
		info    = "build/desktop/windows/info.json"
	)
	templateFiles := map[string]string{
		nfpm:    "name: starter-desktop\ndescription: placeholder\nmaintainer: Starter <starter@example.com>\n",
		desktop: "[Desktop Entry]\nName=Starter\nComment=placeholder\n",
		info:    "{\n\t\"CompanyName\": \"Example\",\n\t\"LegalCopyright\": \"© Example\"\n}\n",
	}
	args := withDetails("--module", "example.com/acme/tasks", "--name", "Acme")

	t.Run("sets them in the three files, after renaming", func(t *testing.T) {
		dir := t.TempDir()
		writeTemplateRecord(t, dir)
		var listed []string
		for rel, content := range templateFiles {
			writeFile(t, dir, rel, content)
			listed = append(listed, rel)
		}

		if _, err := initproject.Run(dir, listed, args, ""); err != nil {
			t.Fatalf("Run error = %v", err)
		}

		for rel, want := range map[string]string{
			nfpm:    "name: acme-desktop\ndescription: \"Tasks for busy teams\"\nmaintainer: \"Acme Ltd <dev@acme.example>\"\n",
			desktop: "[Desktop Entry]\nName=Acme\nComment=Tasks for busy teams\n",
			info:    "{\n\t\"CompanyName\": \"Acme Ltd\",\n\t\"LegalCopyright\": \"© Acme Ltd\"\n}\n",
		} {
			if got := readFile(t, dir, rel); got != want {
				t.Errorf("%s =\n%q\nwant\n%q", rel, got, want)
			}
		}
	})

	t.Run("refuses, touching nothing, when a file lacks its key", func(t *testing.T) {
		dir := t.TempDir()
		writeTemplateRecord(t, dir)
		const gomod = "module github.com/RockLikeAmadeus/go-application-starter\n"
		writeFile(t, dir, "go.mod", gomod)
		for rel, content := range templateFiles {
			writeFile(t, dir, rel, content)
		}
		writeFile(t, dir, desktop, "[Desktop Entry]\nName=Starter\n")

		_, err := initproject.Run(dir, []string{"go.mod", nfpm, desktop, info}, args, "")

		if err == nil || !strings.Contains(err.Error(), "Comment=") {
			t.Errorf("Run error = %v, want one mentioning Comment=", err)
		}
		if got := readFile(t, dir, "go.mod"); got != gomod {
			t.Errorf("go.mod changed to %q", got)
		}
		if got := readFile(t, dir, nfpm); got != templateFiles[nfpm] {
			t.Errorf("nfpm.yaml changed to %q", got)
		}
	})
}
