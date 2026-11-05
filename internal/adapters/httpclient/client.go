// Package httpclient implements the core's driving ports over HTTP, so the
// CLI and TUI can run against a remote server with exactly the same code they
// use against the embedded core.
package httpclient

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// newAPI returns the generated API client for the server at baseURL, with a
// request timeout that opts can replace.
func newAPI(baseURL string, opts ...apiclient.ClientOption) (*apiclient.ClientWithResponses, error) {
	opts = append([]apiclient.ClientOption{
		apiclient.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
	}, opts...)
	return apiclient.NewClientWithResponses(strings.TrimRight(baseURL, "/"), opts...)
}

// toError maps a problem document back to the domain error the server
// started from, so callers can use errors.Is and errors.As regardless of
// transport.
func toError(status int, p *apiclient.Problem) error {
	if p == nil {
		return fmt.Errorf("unexpected HTTP status %d", status)
	}
	typ, detail, field := deref(p.Type), deref(p.Detail), deref(p.Field)
	switch typ {
	case "not_found":
		return domain.ErrNotFound
	case "validation":
		return &domain.ValidationError{Field: field, Message: strings.TrimSpace(strings.TrimPrefix(detail, field))}
	case "limit":
		return &domain.LimitError{Limit: deref(p.Limit), Max: deref(p.Max), ReadOnly: deref(p.ReadOnly)}
	case "feature":
		return &domain.FeatureError{Feature: deref(p.Feature), Name: deref(p.Name), AvailableOn: fromAvailableOn(deref(p.AvailableOn))}
	default:
		return fmt.Errorf("server error (%d): %s", status, detail)
	}
}

// deref returns what p points to, or the zero value when p is nil.
func deref[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}
