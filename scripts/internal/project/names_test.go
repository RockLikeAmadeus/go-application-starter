package project_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

func TestSlug(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr string
	}{
		{name: "Acme", want: "acme"},
		{name: "Acme Tasks", want: "acme-tasks"},
		{name: "  Acme -- Tasks! 2 ", want: "acme-tasks-2"},
		{name: "!!!", wantErr: "name must contain a letter or digit"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := project.Slug(tt.name)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("Slug(%q) error = %v, want %q", tt.name, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Slug(%q) error = %v", tt.name, err)
			}
			if got != tt.want {
				t.Errorf("Slug(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}

func TestEnvPrefix(t *testing.T) {
	for slug, want := range map[string]string{
		"acme":         "ACME_",
		"acme-tasks":   "ACME_TASKS_",
		"acme-tasks-2": "ACME_TASKS_2_",
	} {
		if got := project.EnvPrefix(slug); got != want {
			t.Errorf("EnvPrefix(%q) = %q, want %q", slug, got, want)
		}
	}
}
