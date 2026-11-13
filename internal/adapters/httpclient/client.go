// Package httpclient implements the core's driving ports over HTTP, so the
// CLI and TUI can run against a remote server with exactly the same code they
// use against the embedded core. SessionService also reads the server's
// session endpoint, which no driving port has: it's how a client learns
// whether the server signs people in, and who is signed in.
package httpclient

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

// newClient returns the generated API client for the server at baseURL, with a
// request timeout that opts can replace.
func newClient(baseURL string, opts ...apiclient.ClientOption) (*apiclient.Client, error) {
	opts = append([]apiclient.ClientOption{
		apiclient.WithHTTPClient(&http.Client{Timeout: 10 * time.Second}),
	}, opts...)
	return apiclient.NewClient(strings.TrimRight(baseURL, "/"), opts...)
}

// newAPI returns the client of newClient with its typed responses.
func newAPI(baseURL string, opts ...apiclient.ClientOption) (*apiclient.ClientWithResponses, error) {
	client, err := newClient(baseURL, opts...)
	if err != nil {
		return nil, err
	}
	return &apiclient.ClientWithResponses{ClientInterface: client}, nil
}

// toError maps a problem document back to the domain error the server
// started from, so callers can use errors.Is and errors.As regardless of
// transport.
func toError(status int, p *apiclient.Problem) error {
	if p == nil {
		return fmt.Errorf("The server returned an unexpected response (%d).", status)
	}
	problemType, detail, field := deref(p.Type), deref(p.Detail), deref(p.Field)
	switch problemType {
	case "not_found":
		return domain.ErrNotFound
	case "no_workspace":
		return domain.ErrNoWorkspace
	case "validation":
		return &domain.ValidationError{Field: field, Message: detail}
	case "limit":
		return &domain.LimitError{Limit: deref(p.Limit), Max: deref(p.Max), Plan: deref(p.Plan), IsReadOnly: deref(p.IsReadOnly)}
	case "feature":
		return &domain.FeatureError{Feature: deref(p.Feature), Name: deref(p.Name), AvailableOn: fromAvailableOn(deref(p.AvailableOn))}
	default:
		return fmt.Errorf("The server returned an error (%d): %s", status, detail)
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
