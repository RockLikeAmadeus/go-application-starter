package main_test

import (
	"bytes"
	"errors"
	"slices"
	"strings"
	"testing"

	initproject "github.com/RockLikeAmadeus/go-application-starter/scripts/initproject"
)

func TestFollowUpSkipsPnpmWhenItIsNotInstalled(t *testing.T) {
	tidy := []string{"go", "mod", "tidy"}
	genWeb := []string{"pnpm", "--dir", "web", "gen:api"}
	tests := []struct {
		desc     string
		hasPnpm  bool
		wantRan  [][]string
		wantHint bool
	}{
		{desc: "pnpm installed", hasPnpm: true, wantRan: [][]string{tidy, genWeb}},
		{desc: "pnpm missing", hasPnpm: false, wantRan: [][]string{tidy}, wantHint: true},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			var ran [][]string
			run := func(cmd []string) error { ran = append(ran, cmd); return nil }
			installed := func(name string) bool { return name != "pnpm" || tt.hasPnpm }
			var out bytes.Buffer

			if err := initproject.FollowUp([][]string{tidy, genWeb}, run, installed, &out); err != nil {
				t.Fatalf("FollowUp error = %v", err)
			}

			if !slices.EqualFunc(ran, tt.wantRan, slices.Equal) {
				t.Errorf("ran %q, want %q", ran, tt.wantRan)
			}
			if hint := strings.Contains(out.String(), "run task gen"); hint != tt.wantHint {
				t.Errorf("output %q mentions \"run task gen\" = %v, want %v", out.String(), hint, tt.wantHint)
			}
		})
	}
}

func TestFollowUpTurnsAFailingWebRegenerationIntoAHint(t *testing.T) {
	genWeb := []string{"pnpm", "--dir", "web", "gen:api"}
	run := func([]string) error { return errors.New("ERR_PNPM_UNSUPPORTED_ENGINE") }
	installed := func(string) bool { return true }
	var out bytes.Buffer

	if err := initproject.FollowUp([][]string{genWeb}, run, installed, &out); err != nil {
		t.Fatalf("FollowUp error = %v, want the failure reported as a hint", err)
	}

	for _, want := range []string{"ERR_PNPM_UNSUPPORTED_ENGINE", "run task gen"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("output %q does not mention %q", out.String(), want)
		}
	}
}

func TestFollowUpNamesAFailedCommandOnlyOnce(t *testing.T) {
	genWeb := []string{"pnpm", "--dir", "web", "gen:api"}
	// Errors from main's runner already start with the command line.
	run := func([]string) error { return errors.New("pnpm --dir web gen:api: exit status 1") }
	installed := func(string) bool { return true }
	var out bytes.Buffer

	if err := initproject.FollowUp([][]string{genWeb}, run, installed, &out); err != nil {
		t.Fatalf("FollowUp error = %v", err)
	}

	if n := strings.Count(out.String(), "pnpm --dir web gen:api"); n != 1 {
		t.Errorf("output %q names the command %d times, want once", out.String(), n)
	}
	if !strings.Contains(out.String(), "exit status 1") {
		t.Errorf("output %q does not say why the command failed", out.String())
	}
}
