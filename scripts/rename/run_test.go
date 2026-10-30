package main_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
	rename "github.com/RockLikeAmadeus/go-application-starter/scripts/rename"
)

func TestRunWithDryRunReportsTheChangesAndWritesNothing(t *testing.T) {
	dir := newAcmeProject(t)
	files := map[string]string{
		"README.md":        "# Acme\n\nUnrelated line.\n",
		"cmd/acme/main.go": "package main\n",
	}
	for rel, content := range files {
		writeFile(t, dir, rel, content)
	}
	var out bytes.Buffer

	// A dry run needs no clean tree.
	_, err := rename.Run(dir, []string{"README.md", "cmd/acme/main.go", project.RecordFile},
		[]string{"--name", "Zenith", "--dry-run"}, " M README.md\n", &out)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	for _, want := range []string{"README.md", "- # Acme", "+ # Zenith", "cmd/acme/main.go", "cmd/zenith/main.go"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output does not mention %q:\n%s", want, out.String())
		}
	}
	if strings.Contains(out.String(), "Unrelated line.") {
		t.Errorf("output lists an unchanged line:\n%s", out.String())
	}
	for rel, content := range files {
		if got := readFile(t, dir, rel); got != content {
			t.Errorf("%s = %q, want it unchanged", rel, got)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "cmd", "zenith")); !os.IsNotExist(err) {
		t.Errorf("cmd/zenith exists after a dry run (stat error %v)", err)
	}
	if got, err := project.ReadRecord(dir); err != nil || got != acme {
		t.Errorf("record = %+v, %v; want it unchanged", got, err)
	}
}

func TestRunRenamesTheProject(t *testing.T) {
	dir := newAcmeProject(t)
	writeFile(t, dir, "README.md", "# Acme\n\nSet ACME_STORAGE to choose storage.\n")
	writeFile(t, dir, "cmd/acme/main.go", "// Command acme runs Acme.\npackage main\n")

	_, err := rename.Run(dir, []string{"README.md", "cmd/acme/main.go", project.RecordFile},
		[]string{"--name", "Zenith"}, "", io.Discard)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	if got, want := readFile(t, dir, "README.md"), "# Zenith\n\nSet ZENITH_STORAGE to choose storage.\n"; got != want {
		t.Errorf("README.md = %q, want %q", got, want)
	}
	if got, want := readFile(t, dir, "cmd/zenith/main.go"), "// Command zenith runs Zenith.\npackage main\n"; got != want {
		t.Errorf("cmd/zenith/main.go = %q, want %q", got, want)
	}
	if _, err := os.Stat(filepath.Join(dir, "cmd", "acme")); !os.IsNotExist(err) {
		t.Errorf("cmd/acme still exists (stat error %v)", err)
	}
	want := project.Names{Module: acme.Module, Brand: "Zenith", Slug: "zenith", App: "zenith", EnvPrefix: "ZENITH_"}
	if got, err := project.ReadRecord(dir); err != nil || got != want {
		t.Errorf("record = %+v, %v; want %+v", got, err, want)
	}
}

func TestRunRefusesUncommittedChangesAndChangesNothing(t *testing.T) {
	dir := newAcmeProject(t)
	writeFile(t, dir, "README.md", "# Acme\n")

	_, err := rename.Run(dir, []string{"README.md", project.RecordFile},
		[]string{"--name", "Zenith"}, " M README.md\n", io.Discard)

	if err == nil || !strings.Contains(err.Error(), "commit or stash your changes first") {
		t.Errorf("Run error = %v, want one mentioning %q", err, "commit or stash your changes first")
	}
	if got := readFile(t, dir, "README.md"); got != "# Acme\n" {
		t.Errorf("README.md = %q, want it unchanged", got)
	}
	if got, err := project.ReadRecord(dir); err != nil || got != acme {
		t.Errorf("record = %+v, %v; want it unchanged", got, err)
	}
}

func TestRunWithDryRunListsOnlyFilesARealRunRewrites(t *testing.T) {
	dir := newAcmeProject(t)
	writeFile(t, dir, "go.sum", "example.com/acme/tasks v1.0.0 h1:abc=\n")
	var out bytes.Buffer

	_, err := rename.Run(dir, []string{"go.sum", project.RecordFile},
		[]string{"--name", "Zenith", "--module", "example.com/zenith/notes", "--dry-run"}, "", &out)
	if err != nil {
		t.Fatalf("Run error = %v", err)
	}

	// go.sum is a lock file, and the record is written rather than rewritten.
	for _, skipped := range []string{"go.sum", project.RecordFile} {
		if strings.Contains(out.String(), skipped) {
			t.Errorf("output lists %s:\n%s", skipped, out.String())
		}
	}
}

func TestRunAsksForGoModTidyOnlyWhenTheModuleChanges(t *testing.T) {
	tests := []struct {
		desc string
		args []string
		want [][]string
	}{
		{desc: "module kept", args: []string{"--name", "Zenith"}, want: nil},
		{
			desc: "module changed",
			args: []string{"--name", "Zenith", "--module", "example.com/zenith/notes"},
			want: [][]string{{"go", "mod", "tidy"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			dir := newAcmeProject(t)

			got, err := rename.Run(dir, []string{project.RecordFile}, tt.args, "", io.Discard)
			if err != nil {
				t.Fatalf("Run error = %v", err)
			}

			if !slices.EqualFunc(got, tt.want, slices.Equal) {
				t.Errorf("Run follow-up = %q, want %q", got, tt.want)
			}
		})
	}
}
