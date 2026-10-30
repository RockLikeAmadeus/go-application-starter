package main_test

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMainRunsInAGitRepository(t *testing.T) {
	bin := filepath.Join(t.TempDir(), "initproject")
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	tests := []struct {
		desc      string
		dirty     bool
		wantErr   string // empty for success
		wantGoMod string
	}{
		{
			desc:      "clean tree is made into the new project",
			wantGoMod: "module example.com/acme/tasks\n\ngo 1.22\n",
		},
		{
			desc:      "uncommitted changes are refused",
			dirty:     true,
			wantErr:   "commit or stash your changes first",
			wantGoMod: "module github.com/RockLikeAmadeus/go-application-starter\n\ngo 1.22\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, dir, "go.mod", "module github.com/RockLikeAmadeus/go-application-starter\n\ngo 1.22\n")
			writeTemplateRecord(t, dir)
			git(t, dir, "init", "-q")
			git(t, dir, "add", ".")
			git(t, dir, "-c", "user.name=t", "-c", "user.email=t@example.com", "commit", "-qm", "template")
			if tt.dirty {
				writeFile(t, dir, "notes.txt", "wip\n")
			}

			cmd := exec.Command(bin, "--module", "example.com/acme/tasks", "--name", "Acme")
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()

			if tt.wantErr == "" && err != nil {
				t.Errorf("init failed: %v\n%s", err, out)
			}
			if tt.wantErr != "" && (err == nil || !strings.Contains(string(out), tt.wantErr)) {
				t.Errorf("init error = %v, output %q; want a failure saying %q", err, out, tt.wantErr)
			}
			if got := readFile(t, dir, "go.mod"); got != tt.wantGoMod {
				t.Errorf("go.mod = %q, want %q", got, tt.wantGoMod)
			}
		})
	}
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}
