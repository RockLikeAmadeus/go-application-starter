package main_test

import (
	"testing"

	initproject "github.com/RockLikeAmadeus/go-application-starter/scripts/initproject"
)

var acme = initproject.Names{
	Module:    "example.com/acme/tasks",
	Brand:     "Acme Tasks",
	Slug:      "acme-tasks",
	EnvPrefix: "ACME_TASKS_",
}

func TestRewrite(t *testing.T) {
	tests := []struct {
		desc string
		in   string
		want string
	}{
		{
			desc: "module path",
			in: "module github.com/RockLikeAmadeus/go-application-starter\n" +
				`import "github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"` + "\n",
			want: "module example.com/acme/tasks\n" +
				`import "example.com/acme/tasks/internal/core/domain"` + "\n",
		},
		{
			desc: "app slug and web package",
			in:   "app = 'go-application-starter-staging'\n" + `"name": "starter-web",` + "\n",
			want: "app = 'acme-tasks-staging'\n" + `"name": "acme-tasks-web",` + "\n",
		},
		{
			desc: "environment variables",
			in:   `getenv("STARTER_STORAGE", "jsonfile")` + "\n" + "STARTER_DATA_DIR=./data\n",
			want: `getenv("ACME_TASKS_STORAGE", "jsonfile")` + "\n" + "ACME_TASKS_DATA_DIR=./data\n",
		},
		{
			desc: "brand and command, as whole words only",
			in:   "<title>Starter</title>\ngo run ./cmd/starter tui\nstarters and Starterkit stay\n",
			want: "<title>Acme Tasks</title>\ngo run ./cmd/acme-tasks tui\nstarters and Starterkit stay\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if got := initproject.Rewrite(tt.in, acme); got != tt.want {
				t.Errorf("Rewrite(%q) =\n%q\nwant\n%q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRenamePath(t *testing.T) {
	for path, want := range map[string]string{
		"cmd/starter/main.go":      "cmd/acme-tasks/main.go",
		"cmd/starter/main_test.go": "cmd/acme-tasks/main_test.go",
		"cmd/server/main.go":       "cmd/server/main.go",
		"cmd/starterkit/main.go":   "cmd/starterkit/main.go",
		"docs/starter/notes.md":    "docs/starter/notes.md",
	} {
		if got := initproject.RenamePath(path, acme); got != want {
			t.Errorf("RenamePath(%q) = %q, want %q", path, got, want)
		}
	}
}
