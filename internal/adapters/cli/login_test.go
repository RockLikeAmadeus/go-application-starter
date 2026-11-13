package cli_test

import "testing"

// Signing in stores a token only: it never picks the server, so with none set
// there is nothing to sign in to, however the command line is otherwise set up.
func TestLoginIsRefusedWithoutAServer(t *testing.T) {
	want := "starter login needs a server to sign in to, and none is set. " +
		"Give one with --server <url>, with the STARTER_SERVER_URL environment variable, or with server_url in the settings file. " +
		"Signing in stores a session token only; it never changes which server or data the command line uses, " +
		"so the same server must be set whenever you run starter afterwards."

	out, err := runCLI(t, "--storage", "memory", "login")

	if err == nil || err.Error() != want {
		t.Errorf("login error = %v, want %q", err, want)
	}
	if out != "" {
		t.Errorf("login output = %q, want nothing", out)
	}
}
