// Package httpclient implements the core's driving ports over HTTP, so the
// CLI and TUI can run against a remote server with exactly the same code they
// use against the embedded core.
package httpclient

import (
	"net/http"
	"strings"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
)

// newAPI returns the generated API client for the server at baseURL, with a
// request timeout that opts can replace.
func newAPI(baseURL string, opts ...apiclient.ClientOption) (*apiclient.ClientWithResponses, error) {
	opts = append([]apiclient.ClientOption{
		apiclient.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
	}, opts...)
	return apiclient.NewClientWithResponses(strings.TrimRight(baseURL, "/"), opts...)
}
