package main

import (
	"embed"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	backendAddr := os.Getenv("WHM_BACKEND_URL")
	if backendAddr == "" {
		backendAddr = "http://localhost:3001"
	}

	backendURL, err := url.Parse(backendAddr)
	if err != nil {
		log.Fatalf("invalid WHM_BACKEND_URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	assetHandler := application.AssetFileServerFS(assets)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api") || strings.HasPrefix(r.URL.Path, "/socket.io") {
			proxy.ServeHTTP(w, r)
			return
		}
		assetHandler.ServeHTTP(w, r)
	})

	app := application.New(application.Options{
		Name:        "Wood Hot Monitor",
		Description: "AI-driven multi-source hotspot monitoring",
		Assets: application.AssetOptions{
			Handler: handler,
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
