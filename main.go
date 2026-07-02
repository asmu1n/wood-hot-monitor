package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/wailsapp/wails/v3/pkg/application"

	"wood-hot-monitor/internal/checker"
	"wood-hot-monitor/internal/config"
	"wood-hot-monitor/internal/database"
	"wood-hot-monitor/internal/hotspot"
	"wood-hot-monitor/internal/keyword"
	"wood-hot-monitor/internal/llm"
	"wood-hot-monitor/internal/notification"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	dataDir := defaultDataDir()

	db, err := database.New(dataDir)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	cfgService, err := config.NewService(dataDir)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	llmService := llm.NewService(cfgService, db)

	app := application.New(application.Options{
		Name:        "Wood Hot Monitor",
		Description: "AI-driven multi-source hotspot monitoring",
		Services: []application.Service{
			application.NewService(keyword.NewService(db)),
			application.NewService(hotspot.NewService(db)),
			application.NewService(notification.NewService(db)),
			application.NewService(cfgService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	emitter := func(eventName string, data any) {
		app.Event.EmitEvent(&application.CustomEvent{
			Name: eventName,
			Data: data,
		})
	}

	checkerService := checker.NewService(db, cfgService, llmService, emitter)

	scheduler, err := startScheduler(cfgService, checkerService)
	if err != nil {
		log.Printf("warning: scheduler start failed: %v", err)
	} else {
		defer scheduler.Shutdown()
	}

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

func startScheduler(cfgService *config.Service, checkerService *checker.Service) (gocron.Scheduler, error) {
	cfg, err := cfgService.Get()
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	interval := cfg.CheckInterval
	if interval <= 0 {
		interval = 30
	}

	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("create scheduler: %w", err)
	}

	_, err = s.NewJob(
		gocron.DurationJob(time.Duration(interval)*time.Minute),
		gocron.NewTask(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			if err := checkerService.Run(ctx); err != nil {
				log.Printf("checker run error: %v", err)
			}
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("register job: %w", err)
	}

	s.Start()
	log.Printf("scheduler started: check every %d minutes", interval)
	return s, nil
}

func defaultDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".wood-hot-monitor")
}
