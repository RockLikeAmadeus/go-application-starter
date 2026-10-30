package main_test

import (
	"strings"
	"testing"

	initproject "github.com/RockLikeAmadeus/go-application-starter/scripts/initproject"
)

func TestCheckClean(t *testing.T) {
	tests := []struct {
		desc      string
		porcelain string // output of git status --porcelain
		wantErr   bool
	}{
		{desc: "clean tree", porcelain: "", wantErr: false},
		{desc: "modified file", porcelain: " M go.mod\n", wantErr: true},
		{desc: "untracked file", porcelain: "?? notes.txt\n", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			err := initproject.CheckClean(tt.porcelain)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("CheckClean(%q) = %v, want nil", tt.porcelain, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "commit or stash your changes first") {
				t.Fatalf("CheckClean(%q) = %v, want an error saying to commit or stash", tt.porcelain, err)
			}
		})
	}
}
