//go:build !devsignin

package bootstrap

// isDevSignInBuilt says whether this build has the dev sign-in: only builds
// with the devsignin tag do.
const isDevSignInBuilt = false
