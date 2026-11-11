// Command server runs the HTTP API and serves the web UI.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/httpapi"
	"github.com/RockLikeAmadeus/go-application-starter/internal/bootstrap"
	"github.com/RockLikeAmadeus/go-application-starter/internal/config"
	"github.com/RockLikeAmadeus/go-application-starter/web"
)

func main() {
	if status := runAndReport(config.Env{Getenv: os.Getenv}, os.Stderr); status != 0 {
		os.Exit(status)
	}
}

// runAndReport runs the server until it stops and returns the exit status. An
// error that stopped it goes on errOut (stderr).
func runAndReport(env config.Env, errOut io.Writer) int {
	if err := run(env); err != nil {
		fmt.Fprintln(errOut, "Error:", err)
		return 1
	}
	return 0
}

func run(env config.Env) error {
	cfg, err := loadConfig(env)
	if err != nil {
		return err
	}
	log := bootstrap.NewLogger(cfg, os.Stdout)

	application, err := bootstrap.NewEmbedded(cfg)
	if err != nil {
		return err
	}
	defer application.Close()
	logWarnings(log, application.Warnings)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Port),
		Handler:           httpapi.NewHandler(handlerOptions(application, log)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Info("listening", "addr", srv.Addr, "storage", cfg.Storage, "dataDir", cfg.DataDir, "webUI", web.Assets() != nil)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		log.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return err
		}
	}
	return nil
}

// loadConfig loads the server's configuration: from the settings file
// STARTER_CONFIG_FILE names, if any, under the environment. Unlike the
// local front ends, the server never looks in the user's settings folder,
// so a developer's own settings file doesn't redirect `task dev`'s server.
func loadConfig(env config.Env) (config.Config, error) {
	return config.Load(env, env.Getenv("STARTER_CONFIG_FILE"))
}

// handlerOptions configures the HTTP handler from the wired application.
func handlerOptions(application *bootstrap.App, log *slog.Logger) httpapi.Options {
	return httpapi.Options{
		// sample:start
		Items: application.Items,
		// sample:end
		Entitlements:  application.Entitlements,
		SignIn:        application.SignIn,
		IsDevSignInOn: application.IsDevSignInOn,
		Ready:         application.Ready,
		Warnings:      application.Warnings,
		Logger:        log,
		WebAssets:     web.Assets(),
	}
}

// logWarnings logs each of the app's warnings, one record per warning at
// warn level, so whoever runs the server sees them, not only the web app's
// users.
func logWarnings(log *slog.Logger, warnings []string) {
	for _, w := range warnings {
		log.Warn(w)
	}
}
