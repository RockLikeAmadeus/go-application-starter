// Package workspace provides the ports.WorkspaceResolver implementations.
package workspace

import (
	"context"

	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/ports"
)

// Local is the local edition's resolver. The local edition has a single
// workspace, so every request works in domain.LocalWorkspace.
type Local struct{}

var _ ports.WorkspaceResolver = Local{}

func (Local) CurrentWorkspace(context.Context) (domain.WorkspaceID, error) {
	return domain.LocalWorkspace, nil
}

// FromContext is the resolver for requests whose workspace was put in their
// context with WithID. It fails closed: a request whose workspace is missing
// or empty gets domain.ErrNoWorkspace.
type FromContext struct{}

var _ ports.WorkspaceResolver = FromContext{}

func (FromContext) CurrentWorkspace(ctx context.Context) (domain.WorkspaceID, error) {
	// A missing ID reads as "", so this refuses missing and empty alike.
	id, _ := ctx.Value(idKey{}).(domain.WorkspaceID)
	if id == "" {
		return "", domain.ErrNoWorkspace
	}
	return id, nil
}

// WithID returns a copy of ctx that works in the given workspace.
func WithID(ctx context.Context, id domain.WorkspaceID) context.Context {
	return context.WithValue(ctx, idKey{}, id)
}

// idKey is the context key for the workspace ID. It is unexported, so only
// WithID can set it.
type idKey struct{}
