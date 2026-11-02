// Command starter is the CLI and TUI (`starter tui`) for the application.
// By default it embeds the core; pass --server to use a remote API.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/cli"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	localDataDir := func() (string, error) { return config.LocalDataDir(os.Getenv, os.UserConfigDir) }
	if err := cli.NewRootCmd(bootstrap.New, localDataDir, os.Stdout).ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
