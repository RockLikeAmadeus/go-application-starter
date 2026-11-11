package domain

import "strings"

// Identity is who a sign-in adapter says someone is: the issuer that vouches
// for them, their subject (the issuer's own ID for them), and their name.
// Accounts are found by issuer and subject, never by name.
type Identity struct {
	Issuer  string
	Subject string
	Name    string
}

// NewIdentity builds a valid Identity or returns a *ValidationError. A blank
// issuer or subject is refused, but neither is trimmed: they're compared
// exactly, as an issuer's subjects are opaque.
func NewIdentity(issuer, subject, name string) (Identity, error) {
	if strings.TrimSpace(issuer) == "" {
		return Identity{}, &ValidationError{Field: "issuer", Message: "The sign-in issuer cannot be empty."}
	}
	if strings.TrimSpace(subject) == "" {
		return Identity{}, &ValidationError{Field: "subject", Message: "The sign-in subject cannot be empty."}
	}
	return Identity{Issuer: issuer, Subject: subject, Name: name}, nil
}

// Account is someone who has signed in, and the workspace they work in.
type Account struct {
	ID          string
	Identity    Identity
	WorkspaceID WorkspaceID
}
