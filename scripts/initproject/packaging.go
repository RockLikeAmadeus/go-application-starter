package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
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

// SetInfoJSON sets info.json's CompanyName to the publisher and LegalCopyright
// to "© <publisher>", keeping the layout and every other value, including
// FileDescription and the @VERSION@ placeholders. Values are JSON-escaped
// without HTML escaping, so & < > stay as typed. A missing key is an error.
func SetInfoJSON(text string, o Options) (string, error) {
	text, err := setJSONString(text, "CompanyName", o.Publisher)
	if err != nil {
		return "", err
	}
	return setJSONString(text, "LegalCopyright", "© "+o.Publisher)
}

// setJSONString replaces the string value of the key on its own line, keeping
// the line's indentation and trailing comma, and fails if there's no such line.
func setJSONString(text, key, value string) (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(value); err != nil {
		return "", err
	}
	encoded := strings.TrimSuffix(buf.String(), "\n")
	re := regexp.MustCompile(`(?m)^(\s*"` + regexp.QuoteMeta(key) + `": )"[^\n]*?"(,?)$`)
	if !re.MatchString(text) {
		return "", fmt.Errorf("no %q line", key)
	}
	return re.ReplaceAllString(text, "${1}"+strings.ReplaceAll(encoded, "$", "$$")+"${2}"), nil
}
