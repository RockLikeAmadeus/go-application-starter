package acceptance

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/storage/postgres/postgrestest"
)

var opts = godog.Options{
	Format: "pretty",
	Paths:  []string{"../../features"},
	Strict: true, // undefined or pending steps fail the run
	Output: colors.Colored(os.Stdout),
}

func init() {
	// Enables e.g. `go test ./test/acceptance --godog.tags=@wip --godog.format=progress`.
	godog.BindFlags("godog.", flag.CommandLine, &opts)
}

// TestFeatures runs every feature against every driver. Restrict drivers with
// `-run 'TestFeatures/memory'`.
func TestFeatures(t *testing.T) {
	for _, d := range driversFor(t, os.Getenv(postgrestest.ServerURLVariable)) {
		t.Run(d.name, func(t *testing.T) {
			if d.skipReason != "" {
				t.Skip(d.skipReason)
			}
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

// joinTags requires both Godog tag expressions; either may be empty. Godog's
// "," means "or" and binds more loosely than "&&", so when a driver's tags list alternatives with
// commas, the other expression is joined to each alternative in turn.
func joinTags(a, b string) string {
	if a == "" || b == "" {
		return a + b
	}
	alternatives := strings.Split(b, ",")
	for i, alternative := range alternatives {
		alternatives[i] = a + " && " + alternative
	}
	return strings.Join(alternatives, ",")
}
