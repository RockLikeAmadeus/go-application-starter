package terminal

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/test/terminal/term"
)

// starter is the built starter program the journeys run.
var starter string

// timeout is how long a journey waits for the terminal to show something.
const timeout = 5 * time.Second

// journey holds the state of a single scenario.
type journey struct {
	// home is this scenario's own computer: its home folder, and the folder
	// the programs run in.
	home string
	// dataDir is the data folder the user has chosen, if any.
	dataDir string
	// term is the terminal the program last ran in.
	term *term.Terminal
}

func initializeScenario(sc *godog.ScenarioContext) {
	j := &journey{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		home, err := os.MkdirTemp("", "starter-journey-")
		j.home = home
		return ctx, err
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, err error) (context.Context, error) {
		if j.term != nil {
			_ = j.term.Close()
		}
		_ = os.RemoveAll(j.home)
		return ctx, err
	})

	sc.Step(`^my data is in the folder "([^"]*)"$`, j.myDataIsInTheFolder)
	sc.Step(`^I ask the command line where my data is$`, j.iAskTheCommandLineWhereMyDataIs)
	sc.Step(`^it says my data is in the folder "([^"]*)"$`, j.itSaysMyDataIsInTheFolder)
}

// run starts the starter program with args in a new terminal, on this
// scenario's computer.
func (j *journey) run(args ...string) error {
	cmd := exec.Command(starter, args...)
	cmd.Dir = j.home
	cmd.Env = j.env()
	t, err := term.Start(cmd)
	if err != nil {
		return fmt.Errorf("starting starter %s: %w", strings.Join(args, " "), err)
	}
	j.term = t
	return nil
}

// env is the environment of this scenario's computer: the real one, with its
// own home and settings folder, and none of the user's STARTER_* settings.
func (j *journey) env() []string {
	env := []string{
		"HOME=" + j.home,
		"XDG_CONFIG_HOME=" + filepath.Join(j.home, ".config"),
		"TERM=xterm-256color",
	}
	for _, kv := range os.Environ() {
		switch name, _, _ := strings.Cut(kv, "="); {
		case name == "HOME", name == "XDG_CONFIG_HOME", name == "TERM", strings.HasPrefix(name, "STARTER_"):
		default:
			env = append(env, kv)
		}
	}
	return env
}

func (j *journey) myDataIsInTheFolder(folder string) error {
	j.dataDir = folder
	return nil
}

func (j *journey) iAskTheCommandLineWhereMyDataIs() error {
	return j.run("--data-dir", j.dataDir, "info")
}

func (j *journey) itSaysMyDataIsInTheFolder(folder string) error {
	if err := j.term.WaitFor("Data folder: "+folder, timeout); err != nil {
		return err
	}
	return j.succeeds()
}

// succeeds checks the program ended with a success status.
func (j *journey) succeeds() error {
	status, err := j.term.Wait(timeout)
	if err != nil {
		return err
	}
	if status != 0 {
		return fmt.Errorf("exit status = %d, want 0; the screen shows:\n%s", status, j.term.Screen())
	}
	return nil
}
