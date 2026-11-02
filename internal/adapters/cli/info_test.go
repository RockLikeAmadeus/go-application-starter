package cli_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/jsonfile"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

// runCLI executes the CLI with args, as cmd/starter wires it. args choose
// the data folder, so the local one is never looked up.
func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	return runCLIWith(t, noLocalDataDir(t), args...)
}

// runCLIWith executes the CLI with args, finding the local data folder with
// localDataDir.
func runCLIWith(t *testing.T, localDataDir func() (string, error), args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd := cli.NewRootCmd(bootstrap.New, localDataDir, &out)
	cmd.SetArgs(args)
	err := cmd.ExecuteContext(context.Background())
	return out.String(), err
}

func TestInfoNamesTheDataFolder(t *testing.T) {
	dir := t.TempDir()

	out, err := runCLI(t, "--data-dir", dir, "--storage", "jsonfile", "info")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if !strings.Contains(out, dir) {
		t.Errorf("info output doesn't name the data folder %s:\n%s", dir, out)
	}
}

func TestInfoSaysTheDataIsInUseWhileAnotherProgramHasItOpen(t *testing.T) {
	dir := t.TempDir()
	other, err := jsonfile.Open(dir)
	if err != nil {
		t.Fatalf("open the data folder first: %v", err)
	}
	defer other.Close()

	_, err = runCLI(t, "--data-dir", dir, "--storage", "jsonfile", "info")
	if !errors.Is(err, jsonfile.ErrLocked) {
		t.Errorf("info error = %v, want %v", err, jsonfile.ErrLocked)
	}
}

func TestInfoNamesTheWorkspacesPlan(t *testing.T) {
	out, err := runCLI(t, "--data-dir", t.TempDir(), "--storage", "jsonfile", "info")
	if err != nil {
		t.Fatalf("info: %v", err)
	}
	if !strings.Contains(out, "Plan: "+bootstrap.LocalPlan) {
		t.Errorf("info output doesn't name the plan %q:\n%s", bootstrap.LocalPlan, out)
	}
}

// noLocalDataDir stands in for the local data folder in tests that choose
// their own, failing the test if it is looked up.
func noLocalDataDir(t *testing.T) func() (string, error) {
	return func() (string, error) {
		t.Error("the local data folder was looked up, though the test chose one")
		return "", errors.New("no local data folder in this test")
	}
}

func TestWithoutDataDirCommandsUseTheLocalDataFolder(t *testing.T) {
	t.Setenv("STARTER_DATA_DIR", "")
	local := t.TempDir()
	var opened string
	newApp := func(cfg config.Config) (*bootstrap.App, error) {
		opened = cfg.DataDir
		return bootstrap.New(cfg)
	}
	cmd := cli.NewRootCmd(newApp, func() (string, error) { return local, nil }, &bytes.Buffer{})
	cmd.SetArgs([]string{"--storage", "jsonfile", "info"})

	if err := cmd.ExecuteContext(context.Background()); err != nil {
		t.Fatalf("info: %v", err)
	}
	if opened != local {
		t.Errorf("opened the data folder %q, want the local one %q", opened, local)
	}
}

func TestWithoutTheLocalDataFolderCommandsFailSayingWhy(t *testing.T) {
	errNoHome := errors.New("find the data folder: neither $XDG_CONFIG_HOME nor $HOME are defined")
	_, err := runCLIWith(t, func() (string, error) { return "", errNoHome }, "--storage", "jsonfile", "info")
	if !errors.Is(err, errNoHome) {
		t.Errorf("info error = %v, want %v", err, errNoHome)
	}
}
