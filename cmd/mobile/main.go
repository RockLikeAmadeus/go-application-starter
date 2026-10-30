//go:build android

// Command mobile runs the web UI in an Android app, with the core and its data
// on the phone. The WebView loads the SPA from mobile.Open's Handler and sends
// its API requests over the Wails bridge to Bridge.Serve: no network port is
// opened.
package main

import (
	"os"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"

	"github.com/RockLikeAmadeus/go-application-starter/internal/adapters/mobile"
	"github.com/RockLikeAmadeus/go-application-starter/web"
)

// The app is built as a shared library, so Go never calls main itself: Wails
// calls it when Android creates the app's screen. Android does that again in
// the same process after a theme change, or Back and reopen, but the first app
// is still running (on Android app.Run never returns), so run main only once.
var mainOnce sync.Once

func init() { application.RegisterAndroidMain(func() { mainOnce.Do(main) }) }

func main() {
	// Go's stderr goes nowhere on Android, so report the error to logcat
	// (adb logcat -s Wails).
	if err := run(); err != nil {
		application.DefaultLogger(nil).Error("mobile: " + err.Error())
	}
}

func run() error {
	opened, err := mobile.Open(os.Getenv, application.Android.StoragePath(), web.Assets())
	if err != nil {
		return err
	}
	defer opened.Close()

	app := application.New(application.Options{
		Name:     "Starter",
		Assets:   application.AssetOptions{Handler: opened.Handler},
		Services: []application.Service{application.NewService(&Bridge{opened})},
	})
	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title: "Starter",
		URL:   "/",
	})
	return app.Run()
}

// Bridge is the Wails service the web app calls for every API request, as
// main.Bridge.Serve (see web/src/api/client.ts). Wails names it after this
// package and type, so renaming either breaks the app.
type Bridge struct {
	app *mobile.App
}

// Serve sends one request to the API and returns its reply.
func (b *Bridge) Serve(method, path, body string) mobile.Reply {
	return mobile.Serve(b.app.API, method, path, body)
}
