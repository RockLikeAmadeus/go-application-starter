package acceptance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

var opts = godog.Options{
	Format: "pretty",
	Paths:  []string{"../features"},
	Strict: true,
	Output: colors.Colored(os.Stdout),
}

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opts)
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
	sc.Step(`^I make it mine with:$`, p.makeItMineWith)
	sc.Step(`^I try to make it mine with (description|publisher|email) "(.*)"$`, p.tryWithDetail)

	sc.Step(`^the project's Go module is "([^"]*)"$`, p.goModuleIs)
	sc.Step(`^the app is branded "([^"]*)"$`, p.brandedAs)
	sc.Step(`^the command-line tool is called "([^"]*)"$`, p.commandIs)
	sc.Step(`^settings are read from environment variables starting "([^"]*)"$`, p.envPrefixIs)
	sc.Step(`^no file mentions the template's name$`, p.noTemplateName)
	sc.Step(`^the project has the sample Item capability$`, p.hasSample)
	sc.Step(`^the project has no sample Item capability$`, p.hasNoSample)
	sc.Step(`^the specification has no sample Item capability$`, p.specHasNoSample)
	sc.Step(`^the project builds and its Go tests pass$`, p.buildsAndPasses)
	sc.Step(`^the terminal app keeps its error and selection styles$`, p.keepsTUIStyles)
	sc.Step(`^the web app keeps its error message$`, p.keepsWebErrorMessage)
	sc.Step(`^the project's Go code passes the linters$`, p.passesGoLint)
	sc.Step(`^the init tool and its documentation are gone$`, p.toolIsGone)
	sc.Step(`^the project records module "([^"]*)" and name "([^"]*)"$`, p.records)
	sc.Step(`^the project can still be renamed$`, p.canBeRenamed)
	sc.Step(`^it fails saying "([^"]*)"$`, p.failsSaying)
	sc.Step(`^the project is unchanged$`, p.unchanged)
	sc.Step(`^the Linux packages are described as "([^"]*)"$`, p.linuxDescribedAs)
	sc.Step(`^the Linux packages are maintained by "([^"]*)"$`, p.linuxMaintainedBy)
	sc.Step(`^the app's menu entry is described as "([^"]*)"$`, p.menuEntryDescribedAs)
	sc.Step(`^the Windows program is published by "([^"]*)", with copyright "([^"]*)"$`, p.windowsPublishedBy)
	sc.Step(`^the Windows program is still described as "([^"]*)"$`, p.windowsDescribedAs)
	sc.Step(`^the desktop packages name no placeholder publisher$`, p.noPlaceholderPublisher)
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

// packageDetails are the description, publisher and email the tool is given
// when a step doesn't name them: the tool needs them, but most scenarios are
// about something else.
var packageDetails = map[string]string{
	"description": "Tasks for busy teams",
	"publisher":   "Acme Ltd",
	"email":       "dev@acme.example",
}

// withDetails adds a flag for each detail in packageDetails that set doesn't
// already give, then for each value in set; an empty value leaves its flag out.
func withDetails(set map[string]string) []string {
	var args []string
	for _, key := range []string{"module", "name", "description", "publisher", "email"} {
		value, ok := set[key]
		if !ok {
			value = packageDetails[key]
		}
		if value != "" {
			args = append(args, "--"+key, value)
		}
	}
	return args
}

func (p *project) makeItMine(module, name, removal string) error {
	args := withDetails(map[string]string{"module": module, "name": name})
	if removal != "" {
		args = append(args, "--remove-sample")
	}
	return p.mustRun(args)
}

func (p *project) makeItMineWith(table *godog.Table) error {
	set := map[string]string{}
	for _, row := range table.Rows {
		set[row.Cells[0].Value] = row.Cells[1].Value
	}
	return p.mustRun(withDetails(set))
}

func (p *project) mustRun(args []string) error {
	p.runTool(args)
	if p.err != nil {
		return fmt.Errorf("init failed: %w\n%s", p.err, p.out)
	}
	return nil
}

func (p *project) tryToMakeItMine(module, name string) error {
	return p.tryRun(withDetails(map[string]string{"module": module, "name": name}))
}

func (p *project) tryWithDetail(detail, value string) error {
	return p.tryRun(withDetails(map[string]string{
		"module": "example.com/acme/tasks", "name": "Acme", detail: value,
	}))
}

// tryRun runs the tool after noting the project's state, for a run that may fail.
func (p *project) tryRun(args []string) error {
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

func (p *project) keepsTUIStyles() error {
	for _, style := range []string{"errorStyle", "selectedStyle"} {
		if err := p.allContain(map[string]string{"internal/adapters/tui/styles.go": style}); err != nil {
			return err
		}
	}
	return nil
}

func (p *project) keepsWebErrorMessage() error {
	if !p.exists("web/src/components/ErrorMessage.tsx") {
		return errors.New("web/src/components/ErrorMessage.tsx does not exist")
	}
	return nil
}

func (p *project) passesGoLint() error {
	if out, err := run(p.dir, "golangci-lint", "run", "./..."); err != nil {
		return fmt.Errorf("golangci-lint: %w\n%s", err, firstLines(out, 40))
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

func (p *project) records(module, name string) error {
	data, err := os.ReadFile(filepath.Join(p.dir, "project.json"))
	if err != nil {
		return err
	}
	var got struct{ Module, Name string }
	if err := json.Unmarshal(data, &got); err != nil {
		return err
	}
	if got.Module != module || got.Name != name {
		return fmt.Errorf("project.json records module %q and name %q, want %q and %q", got.Module, got.Name, module, name)
	}
	return nil
}

func (p *project) canBeRenamed() error {
	if !p.exists("scripts/rename/main.go") {
		return errors.New("the rename tool is gone")
	}
	return p.allContain(map[string]string{"Taskfile.yml": "\n  rename:"})
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

const (
	nfpmFile    = "build/desktop/linux/nfpm.yaml"
	desktopFile = "build/desktop/linux/app.desktop"
	infoFile    = "build/desktop/windows/info.json"
)

func (p *project) linuxDescribedAs(description string) error {
	return p.yamlFieldIs(nfpmFile, "description", description)
}

func (p *project) linuxMaintainedBy(maintainer string) error {
	return p.yamlFieldIs(nfpmFile, "maintainer", maintainer)
}

func (p *project) menuEntryDescribedAs(comment string) error {
	return p.allContain(map[string]string{desktopFile: "\nComment=" + comment + "\n"})
}

func (p *project) windowsPublishedBy(company, copyright string) error {
	info, err := p.versionStrings()
	if err != nil {
		return err
	}
	if info["CompanyName"] != company || info["LegalCopyright"] != copyright {
		return fmt.Errorf("%s names company %q and copyright %q, want %q and %q",
			infoFile, info["CompanyName"], info["LegalCopyright"], company, copyright)
	}
	return nil
}

func (p *project) windowsDescribedAs(description string) error {
	info, err := p.versionStrings()
	if err != nil {
		return err
	}
	if got := info["FileDescription"]; got != description {
		return fmt.Errorf("%s's FileDescription is %q, want %q", infoFile, got, description)
	}
	return nil
}

// noPlaceholderPublisher looks for the template's placeholder publisher
// ("Example") and addresses (…@example.com) in the desktop packaging. The app
// ID's com.example. prefix is a separate placeholder (see docs/mobile.md).
func (p *project) noPlaceholderPublisher() error {
	out, err := run(p.dir, "git", "grep", "-I", "-n", "-E",
		"-e", `\bExample\b`, "-e", `@example\.com`, "--", "build/desktop")
	if err == nil {
		return fmt.Errorf("the desktop packages still name a placeholder publisher:\n%s", firstLines(out, 20))
	}
	if out != "" {
		return fmt.Errorf("git grep: %w\n%s", err, out)
	}
	return nil
}

// yamlFieldIs checks a top-level field of a YAML file. The value may be
// quoted, as a YAML or JSON string.
func (p *project) yamlFieldIs(file, key, want string) error {
	text, err := p.read(file)
	if err != nil {
		return err
	}
	for line := range strings.SplitSeq(text, "\n") {
		value, ok := strings.CutPrefix(line, key+":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		if unquoted, err := strconv.Unquote(value); err == nil {
			value = unquoted
		}
		if value != want {
			return fmt.Errorf("%s's %s is %q, want %q", file, key, value, want)
		}
		return nil
	}
	return fmt.Errorf("%s has no %s", file, key)
}

// versionStrings reads the Windows program's version info strings.
func (p *project) versionStrings() (map[string]string, error) {
	data, err := os.ReadFile(filepath.Join(p.dir, infoFile))
	if err != nil {
		return nil, err
	}
	var info struct {
		Info map[string]map[string]string `json:"info"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("%s: %w", infoFile, err)
	}
	return info.Info["0409"], nil
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
