//go:build unix

// Package terminal runs the command line and terminal app journeys in
// test/terminal/features on the built programs, in a terminal, as a user
// would. Business rules are specified in /features (test/acceptance).
package terminal

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opts = godog.Options{
	Format: "pretty",
	Paths:  []string{"features"},
	Strict: true, // undefined or pending steps fail the run
	Output: colors.Colored(os.Stdout),
}

func init() {
	// Enables e.g. `go test ./test/terminal --godog.tags=@wip`.
	godog.BindFlags("godog.", flag.CommandLine, &opts)
}

func TestMain(m *testing.M) {
	flag.Parse()
	bin, err := os.MkdirTemp("", "starter-journeys-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	cliProgram = filepath.Join(bin, "starter")
	serverProgram = filepath.Join(bin, "server")
	for out, pkg := range map[string]string{cliProgram: "../../cmd/starter", serverProgram: "../../cmd/server"} {
		if err := build(out, pkg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			_ = os.RemoveAll(bin)
			os.Exit(1)
		}
	}
	code := m.Run()
	_ = os.RemoveAll(bin)
	os.Exit(code)
}

// build builds the program in pkg to out, as a user would install it.
func build(out, pkg string) error {
	cmd := exec.Command("go", "build", "-o", out, pkg)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("building %s: %w", pkg, err)
	}
	return nil
}

// TestJourneys runs every journey in features.
func TestJourneys(t *testing.T) {
	o := opts
	o.TestingT = t
	suite := godog.TestSuite{
		Name:                "terminal",
		ScenarioInitializer: initializeScenario,
		Options:             &o,
	}
	if suite.Run() != 0 {
		t.Fail()
	}
}
