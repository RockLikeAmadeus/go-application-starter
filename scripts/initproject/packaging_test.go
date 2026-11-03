package main_test

import (
	"strings"
	"testing"

	initproject "github.com/RockLikeAmadeus/go-application-starter/scripts/initproject"
)

func TestSetNFPMDetails(t *testing.T) {
	o := initproject.Options{Description: "Tasks: for busy # teams", Publisher: "Acme Ltd", Email: "dev@acme.example"}
	tests := []struct {
		desc    string
		in      string
		want    string
		wantErr string
	}{
		{
			desc: "sets both lines, quoted, keeping the rest",
			in: "# nfpm config\nname: acme-tasks-desktop\ndescription: Placeholder text\n" +
				"maintainer: Starter <starter@example.com>\ndepends:\n  - libgtk-4-1\n",
			want: "# nfpm config\nname: acme-tasks-desktop\ndescription: \"Tasks: for busy # teams\"\n" +
				"maintainer: \"Acme Ltd <dev@acme.example>\"\ndepends:\n  - libgtk-4-1\n",
		},
		{
			desc:    "no description",
			in:      "name: acme-tasks-desktop\nmaintainer: Starter <starter@example.com>\n",
			wantErr: "no description: line",
		},
		{
			desc:    "no maintainer",
			in:      "name: acme-tasks-desktop\ndescription: Placeholder text\n",
			wantErr: "no maintainer: line",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := initproject.SetNFPMDetails(tt.in, o)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("SetNFPMDetails error = %v, want one mentioning %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("SetNFPMDetails error = %v", err)
			}
			if got != tt.want {
				t.Errorf("SetNFPMDetails(%q) =\n%q\nwant\n%q", tt.in, got, tt.want)
			}
		})
	}
}
