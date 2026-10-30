package main_test

import (
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/scripts/internal/project"
	rename "github.com/RockLikeAmadeus/go-application-starter/scripts/rename"
)

func TestNewNamesKeepsTheCurrentModuleWhenNoneIsGiven(t *testing.T) {
	current := project.Names{Module: "example.com/acme/tasks", Brand: "Acme", Slug: "acme", App: "acme", EnvPrefix: "ACME_"}

	got, err := rename.NewNames(rename.Options{Name: "Zenith Notes"}, current)
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
