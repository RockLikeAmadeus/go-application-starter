//go:build desktop

// Command desktop runs the web UI in a native window, with the core and its
// data on this device. The window loads the SPA and its API in-process from
// desktop.Open's handler: no network port is opened.
package main

import (
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/desktop"
	"github.com/RockLikeAmadeus/go-application-starter/web"
)

func main() {
	if status := desktop.RunAndReport(run, os.Stderr); status != 0 {
		os.Exit(status)
	}
}

func run() error {
	opened, err := desktop.Open(os.Getenv, os.UserConfigDir, web.Assets(), nil)
	if err != nil {
		return err
	}
	defer opened.Close()

	app := application.New(application.Options{
		Name:   "Starter",
		Icon:   opened.Icon,
		Assets: application.AssetOptions{Handler: opened.Handler},
		// GTK 4 has no window icons: Linux takes the icon from the installed
		// .desktop file whose name matches this ID.
		Linux: application.LinuxOptions{ApplicationID: "com.example.starter"},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Starter",
		Width:  1024,
		Height: 768,
		URL:    "/",
	})
	return app.Run()
}
