package main

import (
	"bytes"
	"path"
)

// ShouldRewrite reports whether init should replace the template's names in a
// file. Binary files (those containing a NUL byte, as git decides) are left
// alone, and so are lock files: tools write those, and they are regenerated
// after the rename.
func ShouldRewrite(p string, content []byte) bool {
	switch path.Base(p) {
	case "go.sum", "pnpm-lock.yaml":
		return false
	}
	return !bytes.Contains(content, []byte{0})
}
