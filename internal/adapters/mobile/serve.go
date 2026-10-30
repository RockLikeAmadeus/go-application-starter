// Package mobile wires the mobile app. Wails' Android asset path can only make
// GET requests with no body, so the web app sends its API requests over the
// Wails bridge instead, and Serve runs each one through the HTTP handler. It
// has no Wails dependency, so it is unit-testable; cmd/mobile only adds the
// window and the bridge service.
package mobile

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Reply is the handler's answer to one bridged request.
type Reply struct {
	Status      int    `json:"status"`
	ContentType string `json:"contentType"`
	Body        string `json:"body"`
}

// Serve runs one request through h and returns its reply. A request it can't
// build, such as one with a malformed path, gets a 400 reply without reaching
// h, so a bad call never crashes the bridge.
func Serve(h http.Handler, method, path, body string) Reply {
	req, err := http.NewRequestWithContext(context.Background(), method, path, strings.NewReader(body))
	if err != nil {
		return Reply{Status: http.StatusBadRequest}
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return Reply{
		Status:      rec.Code,
		ContentType: rec.Header().Get("Content-Type"),
		Body:        rec.Body.String(),
	}
}
