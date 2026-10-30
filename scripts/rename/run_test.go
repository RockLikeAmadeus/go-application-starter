package main_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
	rename "github.com/RockLikeAmadeus/go-application-starter/scripts/rename"
)

func TestRunWithDryRunReportsTheChangesAndWritesNothing(t *testing.T) {
	dir := t.TempDir()
	if err := project.WriteRecord(dir, acme); err != nil {
		t.Fatal(err)
	}
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
