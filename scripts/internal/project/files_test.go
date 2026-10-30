package project_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
)

func TestShouldRewrite(t *testing.T) {
	text := []byte("module example.com/acme/tasks\n")
	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	tests := []struct {
		path    string
		content []byte
		want    bool
	}{
		{"go.mod", text, true},
		{"web/src/routes/Layout.tsx", text, true},
		{"web/design-inspiration/mockup.png", png, false},
		{"go.sum", text, false},
		{"tools/go.sum", text, false},
		{"web/pnpm-lock.yaml", text, false},
	}
	for _, tt := range tests {
		if got := project.ShouldRewrite(tt.path, tt.content); got != tt.want {
			t.Errorf("ShouldRewrite(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
