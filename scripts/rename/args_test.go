package main_test

import (
	"strings"
	"testing"

	rename "github.com/RockLikeAmadeus/go-application-starter/scripts/rename"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		desc    string
		args    []string
		want    rename.Options
		wantErr string
	}{
		{desc: "name only", args: []string{"--name", "Zenith Notes"}, want: rename.Options{Name: "Zenith Notes"}},
		{
			desc: "name and module",
			args: []string{"--name", "Zenith", "--module", "example.com/zenith/notes"},
			want: rename.Options{Module: "example.com/zenith/notes", Name: "Zenith"},
		},
		{desc: "dry run", args: []string{"--name", "Zenith", "--dry-run"}, want: rename.Options{Name: "Zenith", IsDryRun: true}},
		{desc: "no name", args: []string{"--module", "example.com/zenith/notes"}, wantErr: "--name is required"},
		{
			desc:    "invalid module path",
			args:    []string{"--name", "Zenith", "--module", "not a module"},
			wantErr: "invalid module path",
		},
		{desc: "unknown flag", args: []string{"--name", "Zenith", "--colour", "red"}, wantErr: "--colour"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := rename.ParseArgs(tt.args)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("ParseArgs(%q) error = %v, want one mentioning %q", tt.args, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseArgs(%q) error = %v", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("ParseArgs(%q) = %+v, want %+v", tt.args, got, tt.want)
			}
		})
	}
}
