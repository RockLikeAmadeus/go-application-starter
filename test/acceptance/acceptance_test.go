package acceptance

import (
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/spf13/pflag"
)

var opts = godog.Options{
	Format: "pretty",
	Paths:  []string{"../../features"},
	Strict: true, // undefined or pending steps fail the run
	Output: colors.Colored(os.Stdout),
}

func init() {
	// Enables e.g. `go test ./test/acceptance --godog.tags=@wip --godog.format=progress`.
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	pflag.Parse()
	os.Exit(m.Run())
}

// TestFeatures runs every feature against every driver. Restrict drivers with
// `-run 'TestFeatures/memory'`.
func TestFeatures(t *testing.T) {
	for _, d := range drivers {
		t.Run(d.name, func(t *testing.T) {
			o := opts
			o.TestingT = t
			o.Tags = joinTags(opts.Tags, d.tags)
			suite := godog.TestSuite{
				Name:                d.name,
				ScenarioInitializer: initializeScenario(d),
				Options:             &o,
			}
			if suite.Run() != 0 {
				t.Fail()
			}
		})
	}
}

// joinTags requires both Godog tag expressions; either may be empty.
func joinTags(a, b string) string {
	if a == "" || b == "" {
		return a + b
	}
	return a + " && " + b
}
