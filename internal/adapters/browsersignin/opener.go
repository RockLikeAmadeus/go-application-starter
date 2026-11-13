package browsersignin

import (
	"os"
	"os/exec"
	"strings"
)

// Open opens address in the user's browser: the program named by the BROWSER
// environment variable (with any arguments it carries), given the address as
// its last argument.
func Open(address string) error {
	fields := strings.Fields(os.Getenv("BROWSER"))
	return exec.Command(fields[0], append(fields[1:], address)...).Run()
}
