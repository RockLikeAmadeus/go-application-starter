package acceptance

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/spf13/pflag"
)

var opts = godog.Options{
	Format: "pretty",
	Paths:  []string{"../features"},
	Strict: true,
	Output: colors.Colored(os.Stdout),
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts)
}

func TestMain(m *testing.M) {
	pflag.Parse()
	os.Exit(m.Run())
}

// TestFeatures runs the init tool against throwaway copies of this template.
// It is slow, so it runs from `task test:init`, not `task test`.
func TestFeatures(t *testing.T) {
	o := opts
	o.TestingT = t
	suite := godog.TestSuite{
		Name:                "init",
		ScenarioInitializer: initializeScenario,
		Options:             &o,
	}
	if suite.Run() != 0 {
		t.Fail()
	}
}

// project is one scenario's throwaway copy of the template and the last run of the tool.
type project struct {
	dir    string
	out    string
	err    error
	before string
}

func initializeScenario(sc *godog.ScenarioContext) {
	p := &project{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		dir, err := os.MkdirTemp("", "initproject-*")
		p.dir = dir
		return ctx, err
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		return ctx, os.RemoveAll(p.dir)
	})

	sc.Step(`^a fresh copy of the template$`, p.freshCopy)
	sc.Step(`^the copy has uncommitted changes$`, p.uncommittedChanges)
	sc.Step(`^I make it mine with module "([^"]*)" and name "([^"]*)"(, removing the sample)?$`, p.makeItMine)
	sc.Step(`^I try to make it mine with module "([^"]*)" and name "([^"]*)"$`, p.tryToMakeItMine)

	sc.Step(`^the project's Go module is "([^"]*)"$`, p.goModuleIs)
	sc.Step(`^the app is branded "([^"]*)"$`, p.brandedAs)
	sc.Step(`^the command-line tool is called "([^"]*)"$`, p.commandIs)
	sc.Step(`^settings are read from environment variables starting "([^"]*)"$`, p.envPrefixIs)
	sc.Step(`^no file mentions the template's name$`, p.noTemplateName)
	sc.Step(`^the project has the sample Item capability$`, p.hasSample)
	sc.Step(`^the project has no sample Item capability$`, p.hasNoSample)
	sc.Step(`^the specification has no sample Item capability$`, p.specHasNoSample)
	sc.Step(`^the project builds and its Go tests pass$`, p.buildsAndPasses)
	sc.Step(`^the init tool and its documentation are gone$`, p.toolIsGone)
	sc.Step(`^it fails saying "([^"]*)"$`, p.failsSaying)
	sc.Step(`^the project is unchanged$`, p.unchanged)
}

// --- arranging -------------------------------------------------------------

// freshCopy copies the template's working tree (tracked and untracked, minus
// ignored files) into a new git repository with everything committed.
func (p *project) freshCopy() error {
	root, err := filepath.Abs("../../..")
	if err != nil {
		return err
	}
	list, err := run(root, "git", "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	if err != nil {
		return err
	}
	for rel := range strings.SplitSeq(strings.TrimRight(list, "\x00"), "\x00") {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if errors.Is(err, os.ErrNotExist) {
			continue // deleted but not yet committed
		}
		if err != nil {
			return err
		}
		dst := filepath.Join(p.dir, rel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return err
		}
	}
	for _, args := range [][]string{
		{"init", "-q"},
		{"add", "-A"},
		{"-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-qm", "template"},
	} {
		if _, err := run(p.dir, "git", args...); err != nil {
			return err
		}
	}
	return nil
}

func (p *project) uncommittedChanges() error {
	f, err := os.OpenFile(filepath.Join(p.dir, "README.md"), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString("\nWork in progress.\n")
	return err
}

// --- acting ----------------------------------------------------------------

func (p *project) makeItMine(module, name, removal string) error {
	args := []string{"--module", module, "--name", name}
	if removal != "" {
		args = append(args, "--remove-sample")
	}
	p.runTool(args)
	if p.err != nil {
		return fmt.Errorf("init failed: %w\n%s", p.err, p.out)
	}
	return nil
}

func (p *project) tryToMakeItMine(module, name string) error {
	var args []string
	if module != "" {
		args = append(args, "--module", module)
	}
	if name != "" {
		args = append(args, "--name", name)
	}
	before, err := p.state()
	if err != nil {
		return err
	}
	p.before = before
	p.runTool(args)
	return nil
}

func (p *project) runTool(args []string) {
	p.out, p.err = run(p.dir, "go", append([]string{"run", "./scripts/initproject"}, args...)...)
}

// --- asserting -------------------------------------------------------------

func (p *project) goModuleIs(want string) error {
	gomod, err := p.read("go.mod")
	if err != nil {
		return err
	}
	first, _, _ := strings.Cut(gomod, "\n")
	if got := strings.TrimPrefix(first, "module "); got != want {
		return fmt.Errorf("go.mod module is %q, want %q", got, want)
	}
	return nil
}

func (p *project) brandedAs(name string) error {
	return p.allContain(map[string]string{
		"web/index.html":            "<title>" + name + "</title>",
		"web/src/routes/Layout.tsx": ">" + name + "<",
	})
}

func (p *project) commandIs(command string) error {
	if !p.exists(filepath.Join("cmd", command, "main.go")) {
		return fmt.Errorf("cmd/%s/main.go does not exist", command)
	}
	if p.exists("cmd/starter") {
		return errors.New("cmd/starter still exists")
	}
	return p.allContain(map[string]string{"Taskfile.yml": "./cmd/" + command})
}

func (p *project) envPrefixIs(prefix string) error {
	return p.allContain(map[string]string{"internal/config/config.go": `"` + prefix + `STORAGE"`})
}

func (p *project) noTemplateName() error {
	out, err := run(p.dir, "git", "grep", "--untracked", "-I", "-i", "-n", "-w", "-E",
		"-e", "starter", "-e", "starter_[a-z0-9_]+", "-e", "RockLikeAmadeus")
	if err == nil {
		return fmt.Errorf("the template's name is still mentioned:\n%s", firstLines(out, 20))
	}
	if out != "" {
		return fmt.Errorf("git grep: %w\n%s", err, out)
	}
	return nil // git grep exits 1, silently, when nothing matches
}

var sampleFiles = []string{
	"internal/core/domain/item.go",
	"features/items.feature",
	"web/src/features/items/ItemsPage.tsx",
}

func (p *project) hasSample() error {
	for _, f := range sampleFiles {
		if !p.exists(f) {
			return fmt.Errorf("%s is missing", f)
		}
	}
	return nil
}

func (p *project) hasNoSample() error {
	for _, f := range sampleFiles {
		if p.exists(f) {
			return fmt.Errorf("%s still exists", f)
		}
	}
	return p.noneContain(map[string]string{"api/openapi.yaml": "/items"})
}

func (p *project) specHasNoSample() error {
	return p.noneContain(map[string]string{"SPEC.md": "Manage items"})
}

func (p *project) buildsAndPasses() error {
	for _, args := range [][]string{
		{"build", "./..."},
		{"test", "-count=1", "./internal/...", "./test/..."},
	} {
		if out, err := run(p.dir, "go", args...); err != nil {
			return fmt.Errorf("go %s: %w\n%s", strings.Join(args, " "), err, firstLines(out, 40))
		}
	}
	return nil
}

func (p *project) toolIsGone() error {
	for _, f := range []string{"scripts/initproject", "docs/template.md"} {
		if p.exists(f) {
			return fmt.Errorf("%s still exists", f)
		}
	}
	return p.noneContain(map[string]string{"Taskfile.yml": "\n  init:", "README.md": "task init"})
}

func (p *project) failsSaying(message string) error {
	if p.err == nil {
		return fmt.Errorf("init succeeded, want it to fail saying %q", message)
	}
	if !strings.Contains(p.out, message) {
		return fmt.Errorf("init failed without saying %q:\n%s", message, p.out)
	}
	return nil
}

func (p *project) unchanged() error {
	after, err := p.state()
	if err != nil {
		return err
	}
	if after != p.before {
		return fmt.Errorf("the project changed:\nbefore:\n%s\nafter:\n%s", p.before, after)
	}
	return nil
}

// --- helpers ---------------------------------------------------------------

// state summarises the working tree, so a run that changes nothing can be detected.
func (p *project) state() (string, error) {
	status, err := run(p.dir, "git", "status", "--porcelain", "--untracked-files=all")
	if err != nil {
		return "", err
	}
	diff, err := run(p.dir, "git", "diff")
	return status + diff, err
}

func (p *project) read(rel string) (string, error) {
	data, err := os.ReadFile(filepath.Join(p.dir, rel))
	return string(data), err
}

func (p *project) exists(rel string) bool {
	_, err := os.Stat(filepath.Join(p.dir, rel))
	return err == nil
}

func (p *project) allContain(want map[string]string) error {
	for file, text := range want {
		got, err := p.read(file)
		if err != nil {
			return err
		}
		if !strings.Contains(got, text) {
			return fmt.Errorf("%s does not contain %q", file, text)
		}
	}
	return nil
}

func (p *project) noneContain(unwanted map[string]string) error {
	for file, text := range unwanted {
		got, err := p.read(file)
		if err != nil {
			return err
		}
		if strings.Contains(got, text) {
			return fmt.Errorf("%s still contains %q", file, text)
		}
	}
	return nil
}

func run(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

func firstLines(s string, n int) string {
	lines := strings.SplitN(s, "\n", n+1)
	if len(lines) > n {
		lines = lines[:n]
	}
	return strings.Join(lines, "\n")
}
