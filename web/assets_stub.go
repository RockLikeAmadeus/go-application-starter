//go:build !embedweb

package web

import "io/fs"

// Assets returns nil when the SPA is not embedded in this build.
func Assets() fs.FS { return nil }
