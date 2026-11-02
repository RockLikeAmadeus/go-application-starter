// Package terminal runs the command line and terminal app journeys in
// test/terminal/features on the built programs, in a terminal, as a user
// would. Business rules are specified in /features (test/acceptance).
package terminal

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/spf13/pflag"
)

var opts = godog.Options{
	Format: "pretty",
	Paths:  []string{"features"},
	Strict: true, // undefined or pending steps fail the run
	Output: colors.Colored(os.Stdout),
}

func init() {
	// Enables e.g. `go test ./test/terminal --godog.tags=@wip`.
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	pflag.Parse()
	os.Exit(m.Run())
}

// TestJourneys runs every journey in features.
func TestJourneys(t *testing.T) {
	o := opts
	o.TestingT = t
	suite := godog.TestSuite{
		Name:                "terminal",
		ScenarioInitializer: func(*godog.ScenarioContext) {},
		Options:             &o,
	}
	if suite.Run() != 0 {
		t.Fail()
	}
}
