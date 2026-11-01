package workspace_test

import (
	"context"
	"errors"
	"testing"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/workspace"
	"github.com/RockLikeAmadeus/go-application-starter/internal/core/domain"
)

func TestLocal_IsAlwaysTheLocalWorkspace(t *testing.T) {
	got, err := workspace.Local{}.CurrentWorkspace(context.Background())
	if err != nil {
		t.Fatalf("CurrentWorkspace: %v", err)
	}
	if got != domain.LocalWorkspace {
		t.Errorf("CurrentWorkspace = %q, want %q", got, domain.LocalWorkspace)
	}
}

func TestFromContext_IsTheWorkspaceSetByWithID(t *testing.T) {
	ctx := workspace.WithID(context.Background(), "ws-a")

	got, err := workspace.FromContext{}.CurrentWorkspace(ctx)
	if err != nil {
		t.Fatalf("CurrentWorkspace: %v", err)
	}
	if got != "ws-a" {
		t.Errorf("CurrentWorkspace = %q, want %q", got, "ws-a")
	}
}

func TestFromContext_WithoutAWorkspaceIsRefused(t *testing.T) {
	_, err := workspace.FromContext{}.CurrentWorkspace(context.Background())
	if !errors.Is(err, domain.ErrNoWorkspace) {
		t.Errorf("err = %v, want ErrNoWorkspace", err)
	}
}

func TestFromContext_WithAnEmptyWorkspaceIsRefused(t *testing.T) {
	ctx := workspace.WithID(context.Background(), "")

	_, err := workspace.FromContext{}.CurrentWorkspace(ctx)
	if !errors.Is(err, domain.ErrNoWorkspace) {
		t.Errorf("err = %v, want ErrNoWorkspace", err)
	}
}
