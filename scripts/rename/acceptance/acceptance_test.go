package acceptance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// TestFeatures runs the rename tool against throwaway copies of this project.
// It is slow, so it runs from `task test:rename`, not `task test`.
func TestFeatures(t *testing.T) {
	o := opts
	o.TestingT = t
	suite := godog.TestSuite{
		Name:                "rename",
		ScenarioInitializer: initializeScenario,
		Options:             &o,
	}
	if suite.Run() != 0 {
		t.Fail()
	}
}

// record is project.json: the names the project currently goes by.
type record struct {
	Module string `json:"module"`
	Name   string `json:"name"`
	Slug   string `json:"slug"`
	App    string `json:"app"`
}

// project is one scenario's throwaway copy of the project, the names it had
// when copied, and the last run of the tool.
type project struct {
	dir    string
	old    record
	out    string
	err    error
	before string
}

func initializeScenario(sc *godog.ScenarioContext) {
	p := &project{}

	sc.Before(func(ctx context.Context, _ *godog.Scenario) (context.Context, error) {
		dir, err := os.MkdirTemp("", "rename-*")
		p.dir = dir
		return ctx, err
	})
	sc.After(func(ctx context.Context, _ *godog.Scenario, _ error) (context.Context, error) {
		return ctx, os.RemoveAll(p.dir)
	})

	sc.Step(`^a copy of the project$`, p.copyProject)
	sc.Step(`^the copy has uncommitted changes$`, p.uncommittedChanges)
	sc.Step(`^I rename it to "([^"]*)"(?: with module "([^"]*)")?$`, p.rename)
	sc.Step(`^I commit the changes$`, p.commit)
	sc.Step(`^I preview renaming it to "([^"]*)"$`, p.preview)
	sc.Step(`^I try to rename it with name "([^"]*)" and module "([^"]*)"$`, p.tryToRename)
	sc.Step(`^I try to rename it to its current name$`, p.tryToRenameToCurrent)

	sc.Step(`^the app is branded "([^"]*)"$`, p.brandedAs)
	sc.Step(`^the command-line tool is called "([^"]*)"$`, p.commandIs)
	sc.Step(`^settings are read from environment variables starting "([^"]*)"$`, p.envPrefixIs)
	sc.Step(`^the web package is called "([^"]*)"$`, p.webPackageIs)
	sc.Step(`^no file mentions the old names$`, p.noOldNames)
	sc.Step(`^no file mentions "([^"]*)"$`, p.noMention)
	sc.Step(`^the project's Go module is unchanged$`, p.moduleUnchanged)
	sc.Step(`^the project's Go module is "([^"]*)"$`, p.goModuleIs)
	sc.Step(`^the project records module "([^"]*)" and name "([^"]*)"$`, p.records)
	sc.Step(`^the project builds and its Go tests pass$`, p.buildsAndPasses)
	sc.Step(`^it lists the changes to "([^"]*)"$`, p.lists)
	sc.Step(`^it lists the command-line tool's move to "([^"]*)"$`, p.lists)
	sc.Step(`^it fails saying "([^"]*)"$`, p.failsSaying)
	sc.Step(`^the project is unchanged$`, p.unchanged)
}

// --- arranging -------------------------------------------------------------

// copyProject copies the project's working tree (tracked and untracked, minus
// ignored files) into a new git repository with everything committed, and
// notes the names it goes by.
func (p *project) copyProject() error {
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
	if _, err := run(p.dir, "git", "init", "-q"); err != nil {
		return err
	}
	if err := p.commit(); err != nil {
		return err
	}
	p.old, err = p.record()
	return err
}

func (p *project) commit() error {
	for _, args := range [][]string{
		{"add", "-A"},
		{"-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-qm", "snapshot"},
	} {
		if out, err := run(p.dir, "git", args...); err != nil {
			return fmt.Errorf("git %s: %w\n%s", args[0], err, out)
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

func (p *project) rename(name, module string) error {
	args := []string{"--name", name}
	if module != "" {
		args = append(args, "--module", module)
	}
	p.runTool(args)
	if p.err != nil {
		return fmt.Errorf("rename failed: %w\n%s", p.err, p.out)
	}
	return nil
}

func (p *project) preview(name string) error {
	if err := p.snapshot(); err != nil {
		return err
	}
	p.runTool([]string{"--name", name, "--dry-run"})
	if p.err != nil {
		return fmt.Errorf("dry run failed: %w\n%s", p.err, p.out)
	}
	return nil
}

func (p *project) tryToRename(name, module string) error {
	var args []string
	if name != "" {
		args = append(args, "--name", name)
	}
	if module != "" {
		args = append(args, "--module", module)
	}
	if err := p.snapshot(); err != nil {
		return err
	}
	p.runTool(args)
	return nil
}

func (p *project) tryToRenameToCurrent() error {
	return p.tryToRename(p.old.Name, "")
}

func (p *project) runTool(args []string) {
	p.out, p.err = run(p.dir, "go", append([]string{"run", "./scripts/rename"}, args...)...)
}

// --- asserting -------------------------------------------------------------

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
	if p.exists(filepath.Join("cmd", p.old.Slug)) {
		return fmt.Errorf("cmd/%s still exists", p.old.Slug)
	}
	return p.allContain(map[string]string{"Taskfile.yml": "./cmd/" + command})
}

func (p *project) envPrefixIs(prefix string) error {
	return p.allContain(map[string]string{"internal/config/config.go": `"` + prefix + `STORAGE"`})
}

func (p *project) webPackageIs(name string) error {
	return p.allContain(map[string]string{"web/package.json": `"name": "` + name + `"`})
}

// noOldNames checks that the old brand, slug, deploy app name, module path and
// environment variable prefix appear nowhere, in any case.
func (p *project) noOldNames() error {
	envPrefix := strings.ToUpper(strings.ReplaceAll(p.old.Slug, "-", "_")) + "_"
	return p.noMatch(
		`\b`+regexp.QuoteMeta(p.old.Name)+`\b`,
		`\b`+regexp.QuoteMeta(p.old.Slug)+`\b`,
		`\b`+regexp.QuoteMeta(p.old.App)+`\b`,
		regexp.QuoteMeta(p.old.Module),
		`\b`+regexp.QuoteMeta(envPrefix)+`[A-Za-z0-9_]+`,
	)
}

func (p *project) noMention(word string) error {
	return p.noMatch(`\b` + regexp.QuoteMeta(word) + `\b`)
}

// noMatch fails if any text file in the copy matches one of patterns, ignoring case.
func (p *project) noMatch(patterns ...string) error {
	re := regexp.MustCompile(`(?i)` + strings.Join(patterns, "|"))
	var found []string
	err := filepath.WalkDir(p.dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil || bytes.IndexByte(data, 0) >= 0 {
			return err
		}
		for i, line := range strings.Split(string(data), "\n") {
			if re.MatchString(line) {
				rel, _ := filepath.Rel(p.dir, path)
				found = append(found, fmt.Sprintf("%s:%d: %s", rel, i+1, line))
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(found) > 0 {
		return fmt.Errorf("the old names are still mentioned:\n%s", firstLines(strings.Join(found, "\n"), 20))
	}
	return nil
}

func (p *project) moduleUnchanged() error {
	return p.goModuleIs(p.old.Module)
}

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

func (p *project) records(module, name string) error {
	got, err := p.record()
	if err != nil {
		return err
	}
	if got.Module != module || got.Name != name {
		return fmt.Errorf("project.json records module %q and name %q, want %q and %q", got.Module, got.Name, module, name)
	}
	return nil
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

func (p *project) lists(text string) error {
	if !strings.Contains(p.out, text) {
		return fmt.Errorf("the preview does not mention %q:\n%s", text, firstLines(p.out, 40))
	}
	return nil
}

func (p *project) failsSaying(message string) error {
	if p.err == nil {
		return fmt.Errorf("rename succeeded, want it to fail saying %q", message)
	}
	if !strings.Contains(p.out, message) {
		return fmt.Errorf("rename failed without saying %q:\n%s", message, p.out)
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

func (p *project) record() (record, error) {
	var r record
	data, err := os.ReadFile(filepath.Join(p.dir, "project.json"))
	if err != nil {
		return r, err
	}
	return r, json.Unmarshal(data, &r)
}

// snapshot notes the working tree's state, so a run that changes nothing can be detected.
func (p *project) snapshot() (err error) {
	p.before, err = p.state()
	return err
}

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
