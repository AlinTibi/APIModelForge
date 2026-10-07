package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()
	messages := windows.DefaultMessages()
	messages.DownloadPage = "API Model Forge requires Microsoft Edge WebView2 Runtime. Press OK to open Microsoft's official download page, or Cancel to exit. Nothing will be installed automatically. Minimum version required: "
	messages.ContactAdmin = "API Model Forge requires Microsoft Edge WebView2 Runtime. Install it from https://developer.microsoft.com/microsoft-edge/webview2/ and restart the application."

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "API Model Forge",
		Width:     1280,
		Height:    820,
		MinWidth:  860,
		MinHeight: 560,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 31, B: 39, A: 255},
		OnStartup:        app.startup,
		Windows:          &windows.Options{Messages: messages},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
