package main

import (
	"fmt"
	"strings"
)

// SetNFPMDetails sets nfpm.yaml's top-level description and maintainer
// ("<publisher> <<email>>") lines, keeping everything else. The values are
// double-quoted, so a colon or # in them is still valid YAML: ParseArgs has
// already refused the quotes and backslashes that would break the quoting. A
// missing key is an error, so a template that lost one fails loudly.
func SetNFPMDetails(text string, o Options) (string, error) {
	text, err := replaceLine(text, "description:", `description: "`+o.Description+`"`)
	if err != nil {
		return "", err
	}
	return replaceLine(text, "maintainer:", `maintainer: "`+o.Publisher+" <"+o.Email+`>"`)
}

// replaceLine replaces every line of text that starts with prefix by line,
// and fails if there's none.
func replaceLine(text, prefix, line string) (string, error) {
	lines := strings.SplitAfter(text, "\n")
	found := false
	for i, l := range lines {
		if strings.HasPrefix(l, prefix) {
			lines[i] = line + "\n"
			found = true
		}
	}
	if !found {
		return "", fmt.Errorf("no %s line", prefix)
	}
	return strings.Join(lines, ""), nil
}

// SetDesktopComment sets the desktop file's Comment line to the description,
// keeping everything else, including localised Comment[xx]= lines. The value
// needs no escaping: ParseArgs has already refused the backslashes and line
// breaks that a desktop entry would read as escapes. A missing key is an
// error, as in SetNFPMDetails.
func SetDesktopComment(text string, o Options) (string, error) {
	return replaceLine(text, "Comment=", "Comment="+o.Description)
}
