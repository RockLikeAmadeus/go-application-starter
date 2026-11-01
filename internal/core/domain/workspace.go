package domain

// WorkspaceID identifies the workspace a piece of data belongs to. Use cases
// get the current request's workspace from a ports.WorkspaceResolver.
type WorkspaceID string

// LocalWorkspace is the local edition's only workspace.
const LocalWorkspace WorkspaceID = "local"
