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
)

func TestRunRefusesBeforeTouchingFiles(t *testing.T) {
	const original = "module github.com/RockLikeAmadeus/go-application-starter\n"
	valid := []string{"--module", "example.com/acme/tasks", "--name", "Acme"}
	tests := []struct {
		desc      string
		args      []string
		porcelain string
		wantErr   string
	}{
		{desc: "missing module", args: []string{"--name", "Acme"}, wantErr: "--module is required"},
		{desc: "invalid module", args: []string{"--module", "not a module", "--name", "Acme"}, wantErr: "invalid module path"},
		{desc: "unusable name", args: []string{"--module", "example.com/acme/tasks", "--name", "!!!"}, wantErr: "name must contain a letter or digit"},
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
	write := func(rel, content string) { writeFile(t, dir, rel, content) }
	read := func(rel string) string { return readFile(t, dir, rel) }
	const goSum = "github.com/RockLikeAmadeus/go-application-starter v0.0.0 h1:x=\n"
	const ignored = "// Starter\n"
	write("go.mod", "module github.com/RockLikeAmadeus/go-application-starter\n")
	write("cmd/starter/main.go", "// Command starter is the Starter CLI.\n")
	write("go.sum", goSum)
	write("web/node_modules/x.js", ignored) // not tracked, so not listed
	files := []string{"go.mod", "cmd/starter/main.go", "go.sum"}

	_, err := initproject.Run(dir, files, []string{"--module", "example.com/acme/tasks", "--name", "Acme Tasks"}, "")
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

// writeFile creates dir/rel (a slash-separated path) with content, making any
// parent directories.
func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// readFile returns the content of dir/rel (a slash-separated path).
func readFile(t *testing.T, dir, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestRunRemovesTheSampleOnlyWhenAsked(t *testing.T) {
	const bootstrap = "package bootstrap\n\n" +
		"// sample:start\n" +
		"var items = 1\n" +
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
			args := append([]string{"--module", "example.com/acme/tasks", "--name", "Acme"}, tt.flags...)

			if _, err := initproject.Run(dir, listed, args, ""); err != nil {
				t.Fatalf("Run error = %v", err)
			}

			for _, rel := range []string{"internal/core/domain/item.go", "web/src/features/items"} {
				_, err := os.Stat(filepath.Join(dir, filepath.FromSlash(rel)))
				if exists := err == nil; exists != tt.wantSample {
					t.Errorf("%s exists = %v, want %v", rel, exists, tt.wantSample)
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

	_, err := initproject.Run(dir, listed, []string{"--module", "example.com/acme/tasks", "--name", "Acme"}, "")
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
		[]string{"--module", "example.com/acme/tasks", "--name", "Acme", "--remove-sample"}, "")
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
			args := append([]string{"--module", "example.com/acme/tasks", "--name", "Acme"}, tt.flags...)

			got, err := initproject.Run(t.TempDir(), nil, args, "")
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
	files := []string{
		"internal/core/app/item_service.go", // the directory's only file
		"internal/core/domain/item.go",
		"internal/core/domain/errors.go", // not sample, so domain/ stays
	}
	for _, rel := range files {
		writeFile(t, dir, rel, "package x\n")
	}
	args := []string{"--module", "example.com/acme/tasks", "--name", "Acme", "--remove-sample"}

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
