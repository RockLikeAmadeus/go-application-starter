package main_test

import (
	"strings"
	"testing"

	initproject "github.com/RockLikeAmadeus/go-application-starter/scripts/initproject"
)

func TestParseArgs(t *testing.T) {
	tests := []struct {
		desc    string
		args    []string
		want    initproject.Options
		wantErr string
	}{
		{
			desc: "module and name",
			args: []string{"--module", "example.com/acme/tasks", "--name", "Acme Tasks"},
			want: initproject.Options{Module: "example.com/acme/tasks", Name: "Acme Tasks"},
		},
		{
			desc: "removing the sample",
			args: []string{"--module", "example.com/acme/tasks", "--name", "Acme", "--remove-sample"},
			want: initproject.Options{Module: "example.com/acme/tasks", Name: "Acme", RemoveSample: true},
		},
		{desc: "no module", args: []string{"--name", "Acme"}, wantErr: "--module is required"},
		{desc: "no name", args: []string{"--module", "example.com/acme/tasks"}, wantErr: "--name is required"},
		{
			desc:    "invalid module path",
			args:    []string{"--module", "not a module", "--name", "Acme"},
			wantErr: "invalid module path",
		},
		{desc: "unknown flag", args: []string{"--colour", "red"}, wantErr: "--colour"},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := initproject.ParseArgs(tt.args)
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
