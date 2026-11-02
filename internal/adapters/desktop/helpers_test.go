package desktop_test

// envOf stands in for os.Getenv, reading from vars.
func envOf(vars map[string]string) func(string) string {
	return func(key string) string { return vars[key] }
}

// configDirAt stands in for os.UserConfigDir, always finding dir.
func configDirAt(dir string) func() (string, error) {
	return func() (string, error) { return dir, nil }
}
