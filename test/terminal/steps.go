package terminal

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cucumber/godog"

	"github.com/RockLikeAmadeus/go-application-starter/test/terminal/term"
)

// starter and server are the built programs the journeys run.
var starter, server string

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
	// serverURL is the address of the server this scenario started, if any.
	serverURL string
	// serverCmd is that server's process, and serverOut what it has printed.
	serverCmd *exec.Cmd
	serverOut bytes.Buffer
}

func initializeScenario(sc *godog.ScenarioContext) {
	j := &journey{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		home, err := os.MkdirTemp("", "starter-journey-")
		j.home = home
		return ctx, err
	})
	// The After hook returns only its own errors: Godog keeps the step's.
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		if j.term != nil {
			_ = j.term.Close()
		}
		j.stopServer()
		_ = os.RemoveAll(j.home)
		return ctx, nil
	})

	sc.Step(`^my data is in the folder "([^"]*)"$`, j.myDataIsInTheFolder)
	sc.Step(`^I ask the command line where my data is$`, j.iAskTheCommandLineWhereMyDataIs)
	sc.Step(`^it says my data is in the folder "([^"]*)"$`, j.itSaysMyDataIsInTheFolder)
	sc.Step(`^a server is running$`, j.aServerIsRunning)
	sc.Step(`^I ask the command line where my data is, connected to the server$`, j.iAskTheCommandLineWhereMyDataIsConnectedToTheServer)
	sc.Step(`^it says my data is on the server$`, j.itSaysMyDataIsOnTheServer)
	sc.Step(`^I give the command line the command "([^"]*)"$`, j.iGiveTheCommandLineTheCommand)
	sc.Step(`^it fails, saying "([^"]*)"$`, j.itFailsSaying)
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
	return j.exitStatus(func(status int) bool { return status == 0 }, "0")
}

// exitStatus waits for the program to end, and checks its exit status is ok
// (described as want), showing the screen if it isn't.
func (j *journey) exitStatus(ok func(status int) bool, want string) error {
	status, err := j.term.Wait(timeout)
	if err != nil {
		return err
	}
	if !ok(status) {
		return fmt.Errorf("exit status = %d, want %s; the screen shows:\n%s", status, want, j.term.Screen())
	}
	return nil
}

func (j *journey) iGiveTheCommandLineTheCommand(command string) error {
	return j.run(command)
}

func (j *journey) itFailsSaying(message string) error {
	if err := j.term.WaitFor(message, timeout); err != nil {
		return err
	}
	return j.exitStatus(func(status int) bool { return status != 0 }, "a failure")
}

// startServer starts the built server for this scenario and sets serverURL.
func (j *journey) startServer() error {
	port, err := freePort()
	if err != nil {
		return err
	}
	cmd := exec.Command(server)
	cmd.Dir = j.home
	cmd.Env = append(j.env(), "STARTER_STORAGE=memory", fmt.Sprintf("PORT=%d", port))
	cmd.Stdout, cmd.Stderr = &j.serverOut, &j.serverOut
	if err := cmd.Start(); err != nil {
		return err
	}
	j.serverCmd = cmd
	j.serverURL = fmt.Sprintf("http://127.0.0.1:%d", port)
	if !ready(j.serverURL, timeout) {
		j.stopServer() // before reading what it printed
		return fmt.Errorf("the server at %s wasn't ready within %s; it printed:\n%s", j.serverURL, timeout, j.serverOut.String())
	}
	return nil
}

// stopServer stops the scenario's server, if it's running.
func (j *journey) stopServer() {
	if j.serverCmd == nil {
		return
	}
	_ = j.serverCmd.Process.Kill()
	_ = j.serverCmd.Wait()
	j.serverCmd = nil
}

// freePort returns a TCP port nothing is listening on now.
func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// ready waits until the server at url says it's ready, and says whether it
// did within timeout.
func ready(url string, timeout time.Duration) bool {
	for deadline := time.Now().Add(timeout); time.Now().Before(deadline); time.Sleep(20 * time.Millisecond) {
		resp, err := http.Get(url + "/readyz")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return true
			}
		}
	}
	return false
}

func (j *journey) aServerIsRunning() error {
	return j.startServer()
}

func (j *journey) iAskTheCommandLineWhereMyDataIsConnectedToTheServer() error {
	return j.run("--server", j.serverURL, "info")
}

func (j *journey) itSaysMyDataIsOnTheServer() error {
	if err := j.term.WaitFor("Server: "+j.serverURL, timeout); err != nil {
		return err
	}
	return j.succeeds()
}
