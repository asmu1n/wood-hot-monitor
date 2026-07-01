package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"

	"wood-hot-monitor/internal/config"
	"wood-hot-monitor/internal/hotspot"
	"wood-hot-monitor/internal/keyword"
	"wood-hot-monitor/internal/notification"
	"wood-hot-monitor/internal/setting"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := application.New(application.Options{
		Name:        "Wood Hot Monitor",
		Description: "AI-driven multi-source hotspot monitoring",
		Services: []application.Service{
			application.NewService(&keyword.Service{}),
			application.NewService(&hotspot.Service{}),
			application.NewService(&notification.Service{}),
			application.NewService(&setting.Service{}),
			application.NewService(&config.Service{}),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Wood Hot Monitor",
		Width:  1280,
		Height: 860,
		URL:    "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
