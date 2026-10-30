package main_test

import (
	"strings"
	"testing"

	initproject "github.com/RockLikeAmadeus/go-application-starter/scripts/initproject"
)

func TestRemoveBlocks(t *testing.T) {
	tests := []struct {
		desc    string
		in      string
		want    string
		wantErr string
	}{
		{
			desc: "Go comments",
			in:   "a := 1\n\t// sample:start\n\tb := 2\n\t// sample:end\nc := 3\n",
			want: "a := 1\nc := 3\n",
		},
		{
			desc: "YAML comments, two blocks",
			in:   "x: 1\n# sample:start\ny: 2\n# sample:end\nz: 3\n# sample:start\nw: 4\n# sample:end\n",
			want: "x: 1\nz: 3\n",
		},
		{
			desc: "HTML and JSX comments",
			in:   "<p>keep</p>\n<!-- sample:start -->\n<p>drop</p>\n{/* sample:end */}\n",
			want: "<p>keep</p>\n",
		},
		{
			desc: "other tags are left alone",
			in:   "// template:start\nkeep\n// template:end\n",
			want: "// template:start\nkeep\n// template:end\n",
		},
		{desc: "start without end", in: "// sample:start\nx\n", wantErr: "line 1: sample:start without sample:end"},
		{desc: "end without start", in: "x\n// sample:end\n", wantErr: "line 2: sample:end without sample:start"},
		{
			desc:    "nested start",
			in:      "// sample:start\n// sample:start\n// sample:end\n",
			wantErr: "line 2: sample:start inside another sample block",
		},
	}
	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := initproject.RemoveBlocks(tt.in, "sample")
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("RemoveBlocks error = %v, want one mentioning %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("RemoveBlocks error = %v", err)
			}
			if got != tt.want {
				t.Errorf("RemoveBlocks(%q) =\n%q\nwant\n%q", tt.in, got, tt.want)
			}
		})
	}
}
