package main_test

import (
	"testing"

	initproject "github.com/RockLikeAmadeus/go-application-starter/scripts/initproject"
)

func TestShouldRewrite(t *testing.T) {
	text := []byte("module github.com/RockLikeAmadeus/go-application-starter\n")
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
		if got := initproject.ShouldRewrite(tt.path, tt.content); got != tt.want {
			t.Errorf("ShouldRewrite(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
