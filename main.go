package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	appchecker "wood-hot-monitor/internal/application/checker"
	apphotspot "wood-hot-monitor/internal/application/hotspot"
	appkeyword "wood-hot-monitor/internal/application/keyword"
	"wood-hot-monitor/internal/core/config"
	"wood-hot-monitor/internal/core/models"
	"wood-hot-monitor/internal/infra/database"
	infraevent "wood-hot-monitor/internal/infra/event"
	infrallm "wood-hot-monitor/internal/infra/llm"
	entpersist "wood-hot-monitor/internal/infra/persistence/ent"
	infrascraper "wood-hot-monitor/internal/infra/scraper"

	"github.com/go-co-op/gocron/v2"
	"github.com/wailsapp/wails/v3/pkg/application"
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

	hotspotRepo := entpersist.NewHotspotRepository(db.Client)
	keywordRepo := entpersist.NewKeywordRepository(db.Client)

	hotspotService := apphotspot.NewService(hotspotRepo)
	keywordService := appkeyword.NewService(keywordRepo)

	app := application.New(application.Options{
		Name:        "Wood Hot Monitor",
		Description: "AI-driven multi-source hotspot monitoring",
		Services: []application.Service{
			application.NewService(keywordService),
			application.NewService(hotspotService),
			application.NewService(cfgService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	notifier := infraevent.NewWailsNotifier(app)
	llmService := infrallm.NewService(cfgService, db.Client)
	scraperService := infrascraper.NewService()
	checkerService := appchecker.NewService(keywordService, cfgService, llmService, hotspotService, scraperService, notifier)
	app.RegisterService(application.NewService(checkerService))

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:  "Wood Hot Monitor",
		Width:  1280,
		Height: 860,
		URL:    "/",
	})

	scheduler, err := startScheduler(cfgService, checkerService)
	if err != nil {
		log.Printf("warning: scheduler start failed: %v", err)
	} else {
		defer scheduler.Shutdown()
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func startScheduler(cfgService *config.Service, checkerService *appchecker.Service) (gocron.Scheduler, error) {
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

	job, err := s.NewJob(
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

	cfgService.SubscribeUpdates(func(ac models.AppConfig) {
		updatedJob, err := s.Update(job.ID(), gocron.DurationJob(time.Duration(ac.CheckInterval)*time.Minute), gocron.NewTask(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			if err := checkerService.Run(ctx); err != nil {
				log.Printf("checker run error: %v", err)
			}
		}))
		if err != nil {
			log.Printf("update job error: %v", err)
		}
		log.Printf("job updated: %v", updatedJob)
	})

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
