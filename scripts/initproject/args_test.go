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
			args: withDetails("--module", "example.com/acme/tasks", "--name", "Acme Tasks"),
			want: wantDetails(initproject.Options{Module: "example.com/acme/tasks", Name: "Acme Tasks"}),
		},
		{
			desc: "removing the sample",
			args: withDetails("--module", "example.com/acme/tasks", "--name", "Acme", "--remove-sample"),
			want: wantDetails(initproject.Options{Module: "example.com/acme/tasks", Name: "Acme", RemoveSample: true}),
		},
		{desc: "no module", args: []string{"--name", "Acme"}, wantErr: "--module is required"},
		{desc: "no name", args: []string{"--module", "example.com/acme/tasks"}, wantErr: "--name is required"},
		{
			desc:    "invalid module path",
			args:    withDetails("--module", "not a module", "--name", "Acme"),
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

func TestParseArgsRequiresEachPackageDetail(t *testing.T) {
	for i := 0; i < len(details); i += 2 {
		flag := details[i]
		t.Run(flag, func(t *testing.T) {
			args := []string{"--module", "example.com/acme/tasks", "--name", "Acme"}
			args = append(args, details[:i]...)
			args = append(args, details[i+2:]...)
			_, err := initproject.ParseArgs(args)
			if want := flag + " is required"; err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("ParseArgs(%q) error = %v, want one mentioning %q", args, err, want)
			}
		})
	}
}
