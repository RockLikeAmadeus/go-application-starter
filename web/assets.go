//go:build embedweb

// Package web exposes the built single-page app (web/dist) to the Go server.
// Build with `-tags embedweb` after `pnpm build` to embed it; without the
// tag the server runs API-only (see assets_stub.go).
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// Assets returns the contents of web/dist.
func Assets() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
