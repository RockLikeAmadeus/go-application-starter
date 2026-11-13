package httpclient

import (
	"context"
	"net/http"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpclient/apiclient"
)

// TokenSource returns the session token to send, or "" for none. It is asked
// on each request, so a token that appears or is forgotten takes effect
// without building a new client. An error fails the request.
type TokenSource func(ctx context.Context) (string, error)

// WithTokenSource makes every request carry the token src returns as a
// bearer token, and none when it returns "".
func WithTokenSource(src TokenSource) apiclient.ClientOption {
	return apiclient.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
		token, err := src(ctx)
		if err != nil {
			return err
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		return nil
	})
}
