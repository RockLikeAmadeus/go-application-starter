package bootstrap

// devSignInBuildTag is the Go build tag that builds in the dev sign-in
// (isDevSignInBuilt). `task dev` and `task run:*` pass it; `task build`
// doesn't, so a release build refuses sign_in = "dev" at start.
const devSignInBuildTag = "devsignin"
