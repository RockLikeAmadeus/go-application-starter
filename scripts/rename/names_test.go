package main_test

import (
	"strings"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
	rename "github.com/RockLikeAmadeus/go-application-starter/scripts/rename"
)

func TestNewNamesKeepsTheCurrentModuleWhenNoneIsGiven(t *testing.T) {
	got, err := rename.NewNames(rename.Options{Name: "Zenith Notes"}, acme)
	if err != nil {
		t.Fatalf("NewNames error = %v", err)
	}

	want := project.Names{
		Module:    "example.com/acme/tasks",
		Brand:     "Zenith Notes",
		Slug:      "zenith-notes",
		App:       "zenith-notes",
		EnvPrefix: "ZENITH_NOTES_",
	}
	if got != want {
		t.Errorf("NewNames = %+v, want %+v", got, want)
	}
}

func TestNewNamesUsesTheModuleWhenOneIsGiven(t *testing.T) {
	got, err := rename.NewNames(rename.Options{Name: "Zenith", Module: "example.com/zenith/notes"}, acme)
	if err != nil {
		t.Fatalf("NewNames error = %v", err)
	}

	if got.Module != "example.com/zenith/notes" {
		t.Errorf("NewNames module = %q, want %q", got.Module, "example.com/zenith/notes")
	}
}

func TestNewNamesRefusesTheCurrentNames(t *testing.T) {
	_, err := rename.NewNames(rename.Options{Name: "Acme"}, acme)

	if err == nil || !strings.Contains(err.Error(), "nothing to rename") {
		t.Errorf("NewNames error = %v, want one mentioning %q", err, "nothing to rename")
	}
}
