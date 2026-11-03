package main_test

import (
	"strings"
	"testing"

	initproject "github.com/RockLikeAmadeus/go-application-starter/scripts/initproject"
)

func TestSetNFPMDetails(t *testing.T) {
	o := initproject.Options{Description: "Tasks: for busy # teams", Publisher: "Acme Ltd", Email: "dev@acme.example"}
	tests := []setterTest{
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
	runSetterTests(t, "SetNFPMDetails", initproject.SetNFPMDetails, o, tests)
}

func TestSetDesktopComment(t *testing.T) {
	o := initproject.Options{Description: "Tasks: for busy # teams & more"}
	tests := []setterTest{
		{
			desc: "sets the line, keeping the rest",
			in:   "[Desktop Entry]\nType=Application\nName=Acme Tasks\nComment=Placeholder text\nComment[de]=Platzhalter\nCategories=Utility;\n",
			want: "[Desktop Entry]\nType=Application\nName=Acme Tasks\nComment=Tasks: for busy # teams & more\nComment[de]=Platzhalter\nCategories=Utility;\n",
		},
		{
			desc:    "no Comment",
			in:      "[Desktop Entry]\nType=Application\nName=Acme Tasks\n",
			wantErr: "no Comment= line",
		},
	}
	runSetterTests(t, "SetDesktopComment", initproject.SetDesktopComment, o, tests)
}

// setterTest is one case for a function that sets lines in a packaging file:
// the text in and either the text wanted or part of the error wanted.
type setterTest struct {
	desc    string
	in      string
	want    string
	wantErr string
}

func runSetterTests(t *testing.T, name string, set func(string, initproject.Options) (string, error), o initproject.Options, tests []setterTest) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := set(tt.in, o)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("%s error = %v, want one mentioning %q", name, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("%s error = %v", name, err)
			}
			if got != tt.want {
				t.Errorf("%s(%q) =\n%q\nwant\n%q", name, tt.in, got, tt.want)
			}
		})
	}
}
