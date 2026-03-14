package main

import (
	"context"
	"embed"
	"os"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"tablepro/internal/connection"
	"tablepro/internal/deeplink"
)

//go:embed all:frontend/dist
var assets embed.FS

var deepLinkHandler *deeplink.DeepLinkHandler

func main() {
	deepLinkHandler = deeplink.NewDeepLinkHandler()

	deepLinkHandler.SetConnectionCallback(func(conn *connection.DatabaseConnection) error {
		runtime.EventsEmit(context.Background(), "deeplink:open-connection", conn)
		return nil
	})

	app := NewApp()

	err := wails.Run(&options.App{
		Title:     "TablePro",
		Width:     1280,
		Height:    720,
		MinWidth:  1024,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.startup(ctx)
			deepLinkHandler.MarkReady()
			handleDeepLinkArgs()
		},
		OnShutdown: app.shutdown,
		Mac: &mac.Options{
			OnUrlOpen: func(url string) {
				deepLinkHandler.Handle(url)
			},
		},
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func handleDeepLinkArgs() {
	args := os.Args[1:]
	for _, arg := range args {
		if strings.HasPrefix(arg, "tablepro://") {
			deepLinkHandler.Handle(arg)
		}
	}
}
